package diff

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/fossology/report-aggregator/internal/model"
)

// DiffResult holds the differences between two reports
type DiffResult struct {
	Added    []model.Component `json:"added"`
	Removed  []model.Component `json:"removed"`
	Modified []ModifiedComponent `json:"modified"`
}

// ModifiedComponent represents a component that changed between versions
type ModifiedComponent struct {
	Name    string   `json:"name"`
	Version string   `json:"version"`
	Changes []Change `json:"changes"` // What fields changed
}

// Change represents a specific field modification
type Change struct {
	Field    string `json:"field"`
	OldValue string `json:"oldValue"`
	NewValue string `json:"newValue"`
}

// ComputeHash computes the SHA-256 hash of a file
func ComputeHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// buildComponentMap maps components by PURL or Name+Version
func buildComponentMap(report *model.ParsedReport) map[string]model.Component {
	compMap := make(map[string]model.Component)
	for _, comp := range report.Components {
		key := comp.PURL
		if key == "" {
			key = strings.ToLower(comp.Name) + "@" + strings.ToLower(comp.Version)
		}
		compMap[key] = comp
	}
	return compMap
}

// DetectChanges compares two parsed reports and finds the differences
func DetectChanges(oldReport, newReport *model.ParsedReport) DiffResult {
	result := DiffResult{}

	oldMap := buildComponentMap(oldReport)
	newMap := buildComponentMap(newReport)

	// Find added and modified
	for key, newComp := range newMap {
		oldComp, exists := oldMap[key]
		if !exists {
			result.Added = append(result.Added, newComp)
			continue
		}

		// Detect modifications
		var changes []Change

		if oldComp.Supplier != newComp.Supplier {
			changes = append(changes, Change{Field: "supplier", OldValue: oldComp.Supplier, NewValue: newComp.Supplier})
		}

		oldLicenses := licensesAsString(oldComp.Licenses)
		newLicenses := licensesAsString(newComp.Licenses)
		if oldLicenses != newLicenses {
			changes = append(changes, Change{Field: "license", OldValue: oldLicenses, NewValue: newLicenses})
		}

		if len(changes) > 0 {
			result.Modified = append(result.Modified, ModifiedComponent{
				Name:    newComp.Name,
				Version: newComp.Version,
				Changes: changes,
			})
		}
	}

	// Find removed
	for key, oldComp := range oldMap {
		if _, exists := newMap[key]; !exists {
			result.Removed = append(result.Removed, oldComp)
		}
	}

	return result
}

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
	sort.Strings(ids)
	return strings.Join(ids, ", ")
}
