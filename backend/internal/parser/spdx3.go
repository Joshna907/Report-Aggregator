package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fossology/report-aggregator/internal/model"
)

// SPDX3Parser parses SPDX 3.0 JSON-LD files
type SPDX3Parser struct{}

func (p *SPDX3Parser) Format() string {
	return "spdx3"
}

// spdx3Doc represents the top-level SPDX 3.0 JSON-LD document
type spdx3Doc struct {
	Context string         `json:"@context"`
	Graph   []spdx3Element `json:"@graph"`
}

// spdx3Element represents a single element in the graph (Package, Software, etc.)
type spdx3Element struct {
	Type             string        `json:"type"`
	SPDXID           string        `json:"spdxId"`
	Name             string        `json:"name"`
	PackageVersion   string        `json:"packageVersion,omitempty"`
	Version          string        `json:"version,omitempty"`
	DownloadLocation string        `json:"downloadLocation,omitempty"`
	LicenseConcluded string        `json:"licenseConcluded,omitempty"`
	LicenseDeclared  string        `json:"licenseDeclared,omitempty"`
	CopyrightText    string        `json:"copyrightText,omitempty"`
	Supplier         string        `json:"supplier,omitempty"`
	Description      string          `json:"description,omitempty"`
	ExternalRefs     []spdx3ExtRef   `json:"externalRefs,omitempty"`
	VerifiedUsing    []spdx3Checksum `json:"verifiedUsing,omitempty"`
	// Relationship fields
	From             string `json:"from,omitempty"`
	To               string `json:"to,omitempty"`
	RelationshipType string `json:"relationshipType,omitempty"`
}

type spdx3Checksum struct {
	Type      string `json:"type"` // e.g., "Hash"
	Algorithm string `json:"algorithm"`
	HashValue string `json:"hashValue"`
}

type spdx3ExtRef struct {
	ReferenceType    string `json:"referenceType"`
	ReferenceLocator string `json:"referenceLocator"`
}

func (p *SPDX3Parser) Parse(filePath string) (*model.ParsedReport, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read SPDX3 file: %w", err)
	}

	var doc spdx3Doc
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("failed to parse SPDX3 JSON-LD: %w", err)
	}

	report := &model.ParsedReport{
		FileName: filepath.Base(filePath),
		Format:   "spdx3",
		ParsedAt: time.Now(),
	}

	provenance := model.Provenance{
		ReportFile: filepath.Base(filePath),
		Format:     "spdx3",
		ParsedAt:   time.Now().Format(time.RFC3339),
	}

	// First pass: build SPDX ID → component name map
	spdxIDToName := make(map[string]string)

	for _, elem := range doc.Graph {
		if elem.Type == "Package" || elem.Type == "Software" {
			if elem.SPDXID != "" && elem.Name != "" {
				spdxIDToName[elem.SPDXID] = elem.Name
			}
		}
	}

	for _, elem := range doc.Graph {
		// Process Package and Software types as components
		if elem.Type == "Package" || elem.Type == "Software" {

			version := elem.PackageVersion
			if version == "" {
				version = elem.Version
			}

			comp := model.Component{
				Name:        elem.Name,
				Version:     version,
				Description: elem.Description,
				DownloadURL: elem.DownloadLocation,
				Supplier:    elem.Supplier,
				Provenance:  []model.Provenance{provenance},
			}

			// Extract PURL from external refs
			for _, ref := range elem.ExternalRefs {
				if ref.ReferenceType == "purl" {
					comp.PURL = ref.ReferenceLocator
					break
				}
			}

			// Extract licenses
			licenseExpr := elem.LicenseConcluded
			if licenseExpr == "" {
				licenseExpr = elem.LicenseDeclared
			}
			if licenseExpr != "" && licenseExpr != "NOASSERTION" {
				licenses := splitLicenseExpression(licenseExpr)
				for _, lic := range licenses {
					comp.Licenses = append(comp.Licenses, model.License{
						ID:         lic,
						Provenance: []model.Provenance{provenance},
					})
				}
			}

			// Extract copyrights
			if elem.CopyrightText != "" && elem.CopyrightText != "NOASSERTION" {
				comp.Copyrights = append(comp.Copyrights, model.Copyright{
					Text:       elem.CopyrightText,
					Provenance: []model.Provenance{provenance},
				})
			}

			// Extract hashes
			for _, hash := range elem.VerifiedUsing {
				comp.Hashes = append(comp.Hashes, model.Hash{
					Algorithm: hash.Algorithm,
					Value:     hash.HashValue,
				})
			}

			report.Components = append(report.Components, comp)
		} else if elem.Type == "Relationship" && elem.From != "" && elem.To != "" {
			srcName := spdxIDToName[elem.From]
			tgtName := spdxIDToName[elem.To]
			if srcName == "" {
				srcName = elem.From
			}
			if tgtName == "" {
				tgtName = elem.To
			}
			relType := elem.RelationshipType
			if relType == "" {
				relType = "RELATED_TO"
			}
			report.Relationships = append(report.Relationships, model.Relationship{
				Source:     srcName,
				Target:     tgtName,
				Type:       relType,
				Provenance: []model.Provenance{provenance},
			})
		}
	}

	return report, nil
}
