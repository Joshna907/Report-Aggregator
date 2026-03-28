package merge

import (
	"fmt"
	"strings"
	"time"

	"github.com/fossology/report-aggregator/internal/model"
)

// Merge takes multiple parsed reports and combines them into one MergeResult
func Merge(reports []*model.ParsedReport) (*model.MergeResult, error) {
	if len(reports) == 0 {
		return nil, fmt.Errorf("no reports to merge")
	}

	result := &model.MergeResult{
		MergedAt:      time.Now().Format(time.RFC3339),
		SourceReports: []string{},
		Components:    []model.Component{},
		Conflicts:     []model.Conflict{},
	}

	// Track source reports
	for _, r := range reports {
		result.SourceReports = append(result.SourceReports, r.FileName)
	}

	// Collect all components from all reports
	type componentKey struct {
		name    string
		version string
	}

	// Index for matching — maps component key to index in result.Components
	purlIndex := make(map[string]int)       // PURL -> index
	nameVerIndex := make(map[componentKey]int) // name+version -> index
	hashIndex := make(map[string]int)        // hash -> index

	for _, report := range reports {
		for _, comp := range report.Components {
			matchIndex := -1

			// Identity Resolution: PURL matching takes precedence (canonical identity)
			if comp.PURL != "" {
				if idx, ok := purlIndex[strings.ToLower(comp.PURL)]; ok {
					matchIndex = idx
				}
			}

			// Heuristic resolution: Fallback to exact Name + Version comparison
			if matchIndex == -1 {
				key := componentKey{
					name:    strings.ToLower(strings.TrimSpace(comp.Name)),
					version: normalizeVersion(comp.Version),
				}
				if idx, ok := nameVerIndex[key]; ok {
					matchIndex = idx
				}
			}

			// Content-based resolution: Match via cryptographic hashes
			if matchIndex == -1 {
				for _, hash := range comp.Hashes {
					hashKey := strings.ToLower(hash.Algorithm + ":" + hash.Value)
					if idx, ok := hashIndex[hashKey]; ok {
						matchIndex = idx
						break
					}
				}
			}

			if matchIndex >= 0 {
				// Component already exists — merge data and detect conflicts
				mergeIntoExisting(&result.Components[matchIndex], &comp, &result.Conflicts)
			} else {
				// New component — add to result
				idx := len(result.Components)
				result.Components = append(result.Components, comp)

				// Update indexes
				if comp.PURL != "" {
					purlIndex[strings.ToLower(comp.PURL)] = idx
				}
				key := componentKey{
					name:    strings.ToLower(strings.TrimSpace(comp.Name)),
					version: normalizeVersion(comp.Version),
				}
				if key.name != "" {
					nameVerIndex[key] = idx
				}
				for _, hash := range comp.Hashes {
					hashKey := strings.ToLower(hash.Algorithm + ":" + hash.Value)
					hashIndex[hashKey] = idx
				}
			}
		}
	}

	// Merge relationships
	type relKey struct {
		source string
		target string
		rtype  string
	}
	relIndex := make(map[relKey]int)

	for _, report := range reports {
		for _, rel := range report.Relationships {
			key := relKey{
				source: strings.ToLower(strings.TrimSpace(rel.Source)),
				target: strings.ToLower(strings.TrimSpace(rel.Target)),
				rtype:  strings.ToUpper(strings.TrimSpace(rel.Type)),
			}

			if idx, ok := relIndex[key]; ok {
				// Relationship exists, append provenance if not already present
				existing := &result.Relationships[idx]
				for _, newProv := range rel.Provenance {
					foundProv := false
					for _, p := range existing.Provenance {
						if p.ReportFile == newProv.ReportFile {
							foundProv = true
							break
						}
					}
					if !foundProv {
						existing.Provenance = append(existing.Provenance, newProv)
					}
				}
			} else {
				// New relationship
				idx := len(result.Relationships)
				result.Relationships = append(result.Relationships, rel)
				relIndex[key] = idx
			}
		}
	}

	// Build summary
	result.Summary = buildSummary(result)

	return result, nil
}

// mergeIntoExisting merges data from a new component into an existing one, detecting conflicts
func mergeIntoExisting(existing *model.Component, incoming *model.Component, conflicts *[]model.Conflict) {
	// Check for license conflicts
	existingLicenses := licensesAsString(existing.Licenses)
	incomingLicenses := licensesAsString(incoming.Licenses)

	if existingLicenses != "" && incomingLicenses != "" && existingLicenses != incomingLicenses {
		*conflicts = append(*conflicts, model.Conflict{
			ID:            int64(len(*conflicts) + 1),
			ComponentName: existing.Name,
			ComponentVer:  existing.Version,
			Field:         "license",
			ValueA:        existingLicenses,
			SourceA:       existing.Provenance[0].ReportFile,
			ValueB:        incomingLicenses,
			SourceB:       incoming.Provenance[0].ReportFile,
		})
	}

	// Check for supplier conflicts
	if existing.Supplier != "" && incoming.Supplier != "" && 
		strings.ToLower(existing.Supplier) != strings.ToLower(incoming.Supplier) {
		*conflicts = append(*conflicts, model.Conflict{
			ID:            int64(len(*conflicts) + 1),
			ComponentName: existing.Name,
			ComponentVer:  existing.Version,
			Field:         "supplier",
			ValueA:        existing.Supplier,
			SourceA:       existing.Provenance[0].ReportFile,
			ValueB:        incoming.Supplier,
			SourceB:       incoming.Provenance[0].ReportFile,
		})
	}

	// Merge provenance for component itself
	for _, inProv := range incoming.Provenance {
		found := false
		for _, exProv := range existing.Provenance {
			if exProv.ReportFile == inProv.ReportFile {
				found = true
				break
			}
		}
		if !found {
			existing.Provenance = append(existing.Provenance, inProv)
		}
	}

	// Merge licenses (union — add new ones)
	for _, inLic := range incoming.Licenses {
		found := false
		for i := range existing.Licenses {
			exLic := &existing.Licenses[i]
			if strings.EqualFold(exLic.ID, inLic.ID) {
				found = true
				// Append provenance
				for _, inP := range inLic.Provenance {
					pFound := false
					for _, exP := range exLic.Provenance {
						if exP.ReportFile == inP.ReportFile {
							pFound = true
							break
						}
					}
					if !pFound {
						exLic.Provenance = append(exLic.Provenance, inP)
					}
				}
				break
			}
		}
		if !found {
			existing.Licenses = append(existing.Licenses, inLic)
		}
	}

	// Merge copyrights (union)
	for _, inCopy := range incoming.Copyrights {
		found := false
		for i := range existing.Copyrights {
			exCopy := &existing.Copyrights[i]
			if exCopy.Text == inCopy.Text {
				found = true
				// Append provenance
				for _, inP := range inCopy.Provenance {
					pFound := false
					for _, exP := range exCopy.Provenance {
						if exP.ReportFile == inP.ReportFile {
							pFound = true
							break
						}
					}
					if !pFound {
						exCopy.Provenance = append(exCopy.Provenance, inP)
					}
				}
				break
			}
		}
		if !found {
			existing.Copyrights = append(existing.Copyrights, inCopy)
		}
	}

	// Merge hashes (union, normalize algorithm names)
	for _, inHash := range incoming.Hashes {
		found := false
		normalizedIn := strings.ToUpper(strings.ReplaceAll(inHash.Algorithm, "-", ""))
		for _, exHash := range existing.Hashes {
			normalizedEx := strings.ToUpper(strings.ReplaceAll(exHash.Algorithm, "-", ""))
			if normalizedIn == normalizedEx && strings.EqualFold(exHash.Value, inHash.Value) {
				found = true
				break
			}
		}
		if !found {
			// Normalize algorithm name to canonical form (e.g. SHA-256)
			inHash.Algorithm = normalizedIn
			existing.Hashes = append(existing.Hashes, inHash)
		}
	}

	// Fill in empty fields from incoming
	if existing.PURL == "" && incoming.PURL != "" {
		existing.PURL = incoming.PURL
	}
	if existing.Supplier == "" && incoming.Supplier != "" {
		existing.Supplier = incoming.Supplier
	}
	if existing.Description == "" && incoming.Description != "" {
		existing.Description = incoming.Description
	}
	if existing.DownloadURL == "" && incoming.DownloadURL != "" {
		existing.DownloadURL = incoming.DownloadURL
	}
}

// normalizeVersion handles version string inconsistencies
// "v4.17.21" and "4.17.21" should match
func normalizeVersion(version string) string {
	v := strings.TrimSpace(strings.ToLower(version))
	v = strings.TrimPrefix(v, "v")
	return v
}

// licensesAsString returns a sorted, comma-separated string of license IDs
func licensesAsString(licenses []model.License) string {
	if len(licenses) == 0 {
		return ""
	}
	var ids []string
	for _, lic := range licenses {
		if lic.ID != "" {
			ids = append(ids, lic.ID)
		}
	}
	return strings.Join(ids, ", ")
}

// buildSummary computes summary statistics for the merge result
func buildSummary(result *model.MergeResult) model.Summary {
	summary := model.Summary{
		TotalComponents:  len(result.Components),
		TotalConflicts:   len(result.Conflicts),
		SourceReportsCount: len(result.SourceReports),
		LicenseBreakdown: make(map[string]int),
		SourceBreakdown:  make(map[string]int),
	}

	uniqueLicenses := make(map[string]bool)

	for _, comp := range result.Components {
		// Count by source
		for _, p := range comp.Provenance {
			summary.SourceBreakdown[p.ReportFile]++
		}

		// Count licenses
		for _, lic := range comp.Licenses {
			id := lic.ID
			if id == "" {
				id = lic.Name
			}
			if id != "" {
				summary.LicenseBreakdown[id]++
				uniqueLicenses[id] = true
			}
		}
	}

	summary.TotalLicenses = len(uniqueLicenses)

	return summary
}
