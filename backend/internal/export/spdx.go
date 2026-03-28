package export

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/fossology/report-aggregator/internal/model"
	spdx_common "github.com/spdx/tools-golang/spdx/v2/common"
	spdx "github.com/spdx/tools-golang/spdx/v2/v2_3"
)

// ToSPDX converts a MergeResult into a valid SPDX 2.3 JSON document
func ToSPDX(result *model.MergeResult) ([]byte, error) {
	doc := &spdx.Document{
		SPDXVersion:       "SPDX-2.3",
		DataLicense:       "CC0-1.0",
		SPDXIdentifier:    "SPDXRef-DOCUMENT",
		DocumentName:      "FOSSology-Aggregated-Report",
		DocumentNamespace: fmt.Sprintf("http://spdx.org/spdxdocs/fossology-aggregated-%d", time.Now().Unix()),
		CreationInfo: &spdx.CreationInfo{
			Creators: []spdx_common.Creator{
				{
					CreatorType: "Tool",
					Creator:     "FOSSology-Report-Aggregator",
				},
			},
			Created: time.Now().UTC().Format(time.RFC3339),
		},
	}

	for i, comp := range result.Components {
		pkg := &spdx.Package{
			PackageSPDXIdentifier: spdx_common.ElementID(fmt.Sprintf("Package-%d", i)),
			PackageName:           comp.Name,
			PackageVersion:        comp.Version,
			PackageDescription:    comp.Description,
			PackageDownloadLocation: comp.DownloadURL,
		}

		if comp.Supplier != "" {
			pkg.PackageSupplier = &spdx_common.Supplier{
				SupplierType: "Organization",
				Supplier:     comp.Supplier,
			}
		}

		if comp.DownloadURL == "" {
			pkg.PackageDownloadLocation = "NOASSERTION"
		}

		// Handle PURL
		if comp.PURL != "" {
			pkg.PackageExternalReferences = []*spdx.PackageExternalReference{
				{
					Category: spdx_common.CategoryPackageManager,
					RefType:  "purl",
					Locator:  comp.PURL,
				},
			}
		}

		// Handle Licenses
		if len(comp.Licenses) > 0 {
			var licIDs []string
			for _, lic := range comp.Licenses {
				if lic.ID != "" {
					licIDs = append(licIDs, lic.ID)
				}
			}
			if len(licIDs) > 0 {
				// Very basic joining, technically should build proper SPDX expression
				pkg.PackageLicenseConcluded = licIDs[0]
				for j := 1; j < len(licIDs); j++ {
					pkg.PackageLicenseConcluded += " AND " + licIDs[j]
				}
			} else {
				pkg.PackageLicenseConcluded = "NOASSERTION"
			}
		} else {
			pkg.PackageLicenseConcluded = "NOASSERTION"
		}
		pkg.PackageLicenseDeclared = pkg.PackageLicenseConcluded

		// Handle Hashes
		for _, hash := range comp.Hashes {
			algo := spdx_common.ChecksumAlgorithm(hash.Algorithm)
			pkg.PackageChecksums = append(pkg.PackageChecksums, spdx_common.Checksum{
				Algorithm: algo,
				Value:     hash.Value,
			})
		}

		// Handle Copyrights
		if len(comp.Copyrights) > 0 {
			pkg.PackageCopyrightText = comp.Copyrights[0].Text
		} else {
			pkg.PackageCopyrightText = "NOASSERTION"
		}

		doc.Packages = append(doc.Packages, pkg)
	}

	// Handle Relationships
	for _, rel := range result.Relationships {
		doc.Relationships = append(doc.Relationships, &spdx.Relationship{
			RefA: spdx_common.DocElementID{
				ElementRefID: spdx_common.ElementID(rel.Source),
			},
			RefB: spdx_common.DocElementID{
				ElementRefID: spdx_common.ElementID(rel.Target),
			},
			Relationship: rel.Type,
		})
	}

	return json.MarshalIndent(doc, "", "  ")
}
