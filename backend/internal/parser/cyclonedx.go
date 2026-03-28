package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/CycloneDX/cyclonedx-go"
	"github.com/fossology/report-aggregator/internal/model"
)

// CycloneDXParser parses CycloneDX JSON files
type CycloneDXParser struct{}

func (p *CycloneDXParser) Format() string {
	return "cyclonedx"
}

func (p *CycloneDXParser) Parse(filePath string) (*model.ParsedReport, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CycloneDX file: %w", err)
	}
	defer file.Close()

	bom := cyclonedx.NewBOM()
	decoder := cyclonedx.NewBOMDecoder(file, cyclonedx.BOMFileFormatJSON)
	if err := decoder.Decode(bom); err != nil {
		return nil, fmt.Errorf("failed to decode CycloneDX JSON: %w", err)
	}

	report := &model.ParsedReport{
		FileName: filepath.Base(filePath),
		Format:   "cyclonedx",
		ParsedAt: time.Now(),
	}

	provenance := model.Provenance{
		ReportFile: filepath.Base(filePath),
		Format:     "cyclonedx",
		ParsedAt:   time.Now().Format(time.RFC3339),
	}

	// Track BOM ref to component name for relationships
	refToName := make(map[string]string)

	if bom.Components != nil {
		for _, comp := range *bom.Components {
			refToName[comp.BOMRef] = comp.Name

			// Map core metadata to UCM
			c := model.Component{
				Name:        comp.Name,
				Version:     comp.Version,
				Description: comp.Description,
				PURL:        comp.PackageURL,
				Provenance:  []model.Provenance{provenance},
			}

			if comp.Supplier != nil {
				c.Supplier = comp.Supplier.Name
			}

			if comp.Licenses != nil {
				for _, lic := range *comp.Licenses {
					if lic.License != nil {
						c.Licenses = append(c.Licenses, model.License{
							ID:         lic.License.ID,
							Name:       lic.License.Name,
							Provenance: []model.Provenance{provenance},
						})
					} else if lic.Expression != "" {
						// Simple split for expression
						licenses := splitLicenseExpression(lic.Expression)
						for _, l := range licenses {
							c.Licenses = append(c.Licenses, model.License{
								ID:         l,
								Provenance: []model.Provenance{provenance},
							})
						}
					}
				}
			}

			if comp.Hashes != nil {
				for _, h := range *comp.Hashes {
					c.Hashes = append(c.Hashes, model.Hash{
						Algorithm: string(h.Algorithm),
						Value:     h.Value,
					})
				}
			}

			report.Components = append(report.Components, c)
		}
	}

	if bom.Dependencies != nil {
		for _, dep := range *bom.Dependencies {
			srcName, okSrc := refToName[dep.Ref]
			if !okSrc {
				continue
			}

			if dep.Dependencies != nil {
				for _, subDepRef := range *dep.Dependencies {
					tgtName, okTgt := refToName[subDepRef]
					if okTgt {
						report.Relationships = append(report.Relationships, model.Relationship{
							Source:     srcName,
							Target:     tgtName,
							Type:       "DEPENDS_ON",
							Provenance: []model.Provenance{provenance},
						})
					}
				}
			}
		}
	}

	return report, nil
}