package export

import (
	"bytes"
	"fmt"
	"time"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/fossology/report-aggregator/internal/model"
)

// ToCycloneDX converts a MergeResult into a valid CycloneDX JSON document
func ToCycloneDX(result *model.MergeResult) ([]byte, error) {
	bom := cdx.NewBOM()
	bom.SerialNumber = fmt.Sprintf("urn:uuid:fossology-aggregator-%d", time.Now().Unix())
	bom.Metadata = &cdx.Metadata{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Tools: &cdx.ToolsChoice{
			Components: &[]cdx.Component{
				{
					Type:    cdx.ComponentTypeApplication,
					Name:    "FOSSology Report Aggregator",
					Version: "1.0.0",
				},
			},
		},
	}

	var components []cdx.Component

	for _, comp := range result.Components {
		cdxComp := cdx.Component{
			Type:        cdx.ComponentTypeLibrary,
			Name:        comp.Name,
			Version:     comp.Version,
			Description: comp.Description,
			PackageURL:  comp.PURL,
		}

		if comp.Supplier != "" {
			cdxComp.Supplier = &cdx.OrganizationalEntity{
				Name: comp.Supplier,
			}
		}

		// Licenses
		if len(comp.Licenses) > 0 {
			var cdxLicenses cdx.Licenses
			for _, lic := range comp.Licenses {
				cdxLicenses = append(cdxLicenses, cdx.LicenseChoice{
					License: &cdx.License{
						ID:   lic.ID,
						Name: lic.Name,
					},
				})
			}
			cdxComp.Licenses = &cdxLicenses
		}

		// Hashes
		if len(comp.Hashes) > 0 {
			var cdxHashes []cdx.Hash
			for _, hash := range comp.Hashes {
				cdxHashes = append(cdxHashes, cdx.Hash{
					Algorithm: cdx.HashAlgorithm(hash.Algorithm),
					Value:     hash.Value,
				})
			}
			cdxComp.Hashes = &cdxHashes
		}

		// Copyrights
		if len(comp.Copyrights) > 0 {
			cdxComp.Copyright = comp.Copyrights[0].Text
		}

		components = append(components, cdxComp)
	}

	bom.Components = &components

	// Handle Relationships as Dependencies
	if len(result.Relationships) > 0 {
		var dependencies []cdx.Dependency
		for _, rel := range result.Relationships {
			dependencies = append(dependencies, cdx.Dependency{
				Ref: rel.Source,
				Dependencies: &[]string{
					rel.Target,
				},
			})
		}
		bom.Dependencies = &dependencies
	}

	// Encode to JSON string
	var buf bytes.Buffer
	encoder := cdx.NewBOMEncoder(&buf, cdx.BOMFileFormatJSON)
	encoder.SetPretty(true)

	if err := encoder.Encode(bom); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
