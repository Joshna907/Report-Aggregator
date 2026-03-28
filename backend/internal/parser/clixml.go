package parser

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"time"

	"github.com/fossology/report-aggregator/internal/model"
)

// CLIXML represents the subset of CLIXML tags we care about
type CLIXML struct {
	Components []struct {
		Name     string `xml:"Name"`
		Version  string `xml:"Version"`
		License  string `xml:"License"`
		Supplier string `xml:"Supplier"`
		PURL     string `xml:"PURL"`
		Hash     string `xml:"Hash"`
	} `xml:"Component"`
}


// ParseCLIXML parses a CLIXML report
func ParseCLIXML(path string) (*model.ParsedReport, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var raw CLIXML
	if err := xml.NewDecoder(file).Decode(&raw); err != nil {
		return nil, err
	}

	report := &model.ParsedReport{
		FileName: filepath.Base(path),
		Format:   "clixml",
		ParsedAt: time.Now(),
	}

	for _, c := range raw.Components {
		comp := model.Component{
			Name:     c.Name,
			Version:  c.Version,
			Supplier: c.Supplier,
			PURL:     c.PURL,
			Licenses: []model.License{{
				ID: c.License,
				Provenance: []model.Provenance{{
					ReportFile: filepath.Base(path),
					Format:     "clixml",
					ParsedAt:   time.Now().Format(time.RFC3339),
				}},
			}},
			Provenance: []model.Provenance{{
				ReportFile: filepath.Base(path),
				Format:     "clixml",
				ParsedAt:   time.Now().Format(time.RFC3339),
			}},
		}

		if c.Hash != "" {
			comp.Hashes = append(comp.Hashes, model.Hash{
				Algorithm: "SHA-1", // Assume SHA-1 for CLIXML legacy
				Value:     c.Hash,
			})
		}

		report.Components = append(report.Components, comp)
	}

	return report, nil
}

// CLIXMLParser implements the Parser interface
type CLIXMLParser struct{}

func (p *CLIXMLParser) Parse(path string) (*model.ParsedReport, error) {
	return ParseCLIXML(path)
}

func (p *CLIXMLParser) Format() string {
	return "clixml"
}

