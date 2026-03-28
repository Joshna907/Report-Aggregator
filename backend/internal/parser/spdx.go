package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fossology/report-aggregator/internal/model"
	"github.com/spdx/tools-golang/json"
)

// SPDXParser parses SPDX 2.3 files (JSON format)
type SPDXParser struct{}

func (p *SPDXParser) Format() string {
	return "spdx"
}

func (p *SPDXParser) Parse(filePath string) (*model.ParsedReport, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open SPDX file: %w", err)
	}
	defer file.Close()

	doc, err := json.Read(file)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SPDX JSON: %w", err)
	}

	report := &model.ParsedReport{
		FileName: filepath.Base(filePath),
		Format:   "spdx",
		ParsedAt: time.Now(),
	}

	provenance := model.Provenance{
		ReportFile: filepath.Base(filePath),
		Format:     "spdx",
		ParsedAt:   time.Now().Format(time.RFC3339),
	}

	idToName := make(map[string]string)

	for _, pkg := range doc.Packages {
		idToName[string(pkg.PackageSPDXIdentifier)] = pkg.PackageName
		
		component := model.Component{
			Name:        pkg.PackageName,
			Version:     pkg.PackageVersion,
			Description: pkg.PackageDescription,
			DownloadURL: pkg.PackageDownloadLocation,
			Provenance:  []model.Provenance{provenance},
		}

		// Normalize supplier data
		if pkg.PackageSupplier != nil {
			component.Supplier = pkg.PackageSupplier.Supplier
		}

		// Resolve PURL from external references (highest priority identity)
		if pkg.PackageExternalReferences != nil {
			for _, ref := range pkg.PackageExternalReferences {
				if ref.RefType == "purl" {
					component.PURL = ref.Locator
					break
				}
			}
		}

		// Extract licenses
		if pkg.PackageLicenseConcluded != "" && pkg.PackageLicenseConcluded != "NOASSERTION" {
			licenses := splitLicenseExpression(pkg.PackageLicenseConcluded)
			for _, lic := range licenses {
				component.Licenses = append(component.Licenses, model.License{
					ID:         lic,
					Provenance: []model.Provenance{provenance},
				})
			}
		}

		// Extract hashes/checksums
		if pkg.PackageChecksums != nil {
			for _, checksum := range pkg.PackageChecksums {
				component.Hashes = append(component.Hashes, model.Hash{
					Algorithm: string(checksum.Algorithm),
					Value:     checksum.Value,
				})
			}

		}

		// Extract copyrights
		if pkg.PackageCopyrightText != "" && pkg.PackageCopyrightText != "NOASSERTION" {
			component.Copyrights = append(component.Copyrights, model.Copyright{
				Text:       pkg.PackageCopyrightText,
				Provenance: []model.Provenance{provenance},
			})
		}

		report.Components = append(report.Components, component)
	}

	for _, rel := range doc.Relationships {
		srcID := string(rel.RefA.ElementRefID)
		tgtID := string(rel.RefB.ElementRefID)

		srcName, okSrc := idToName[srcID]
		tgtName, okTgt := idToName[tgtID]

		// Only add relationship if we know both components
		if okSrc && okTgt {
			report.Relationships = append(report.Relationships, model.Relationship{
				Source:     srcName,
				Target:     tgtName,
				Type:       rel.Relationship,
				Provenance: []model.Provenance{provenance},
			})
		}
	}

	return report, nil
}

// splitLicenseExpression splits "MIT AND Apache-2.0" into ["MIT", "Apache-2.0"]
func splitLicenseExpression(expr string) []string {
	expr = strings.ReplaceAll(expr, "(", "")
	expr = strings.ReplaceAll(expr, ")", "")

	parts := strings.FieldsFunc(expr, func(r rune) bool {
		return r == ' '
	})

	var licenses []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "AND" && part != "OR" && part != "" {
			licenses = append(licenses, part)
		}
	}
	return licenses
}
