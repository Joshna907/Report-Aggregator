package parser

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fossology/report-aggregator/internal/model"
)

// ParseReadMeOSS parses a standard ReadMeOSS text file
func ParseReadMeOSS(path string) (*model.ParsedReport, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	report := &model.ParsedReport{
		FileName: filepath.Base(path),
		Format:   "readmeoss",
		ParsedAt: time.Now(),
	}

	scanner := bufio.NewScanner(file)
	var currentComp *model.Component

	for scanner.Scan() {
		line := scanner.Text()

		// Very simple heuristic parser for ReadMeOSS text format:
		// Package: name  (or Component: name)
		// Version: version
		// License: ID

		if strings.HasPrefix(line, "Package:") {
			if currentComp != nil {
				report.Components = append(report.Components, *currentComp)
			}
			name := strings.TrimSpace(strings.TrimPrefix(line, "Package:"))
			currentComp = &model.Component{
				Name: name,
				Provenance: []model.Provenance{{
					ReportFile: filepath.Base(path),
					Format:     "readmeoss",
					ParsedAt:   time.Now().Format(time.RFC3339),
				}},
			}
		} else if strings.HasPrefix(line, "Component:") {
			if currentComp != nil {
				report.Components = append(report.Components, *currentComp)
			}
			name := strings.TrimSpace(strings.TrimPrefix(line, "Component:"))
			currentComp = &model.Component{
				Name: name,
				Provenance: []model.Provenance{{
					ReportFile: filepath.Base(path),
					Format:     "readmeoss",
					ParsedAt:   time.Now().Format(time.RFC3339),
				}},
			}
		} else if currentComp != nil {
			if strings.HasPrefix(line, "Version:") {
				currentComp.Version = strings.TrimSpace(strings.TrimPrefix(line, "Version:"))
			} else if strings.HasPrefix(line, "License:") {
				licID := strings.TrimSpace(strings.TrimPrefix(line, "License:"))
				currentComp.Licenses = append(currentComp.Licenses, model.License{
					ID:         licID,
					Provenance: currentComp.Provenance,
				})
			} else if strings.HasPrefix(line, "PURL:") {
				currentComp.PURL = strings.TrimSpace(strings.TrimPrefix(line, "PURL:"))
			} else if strings.HasPrefix(line, "Supplier:") {
				currentComp.Supplier = strings.TrimSpace(strings.TrimPrefix(line, "Supplier:"))
			}
		}
	}

	if currentComp != nil {
		report.Components = append(report.Components, *currentComp)
	}

	return report, nil
}

// ReadMeOSSParser implements the Parser interface
type ReadMeOSSParser struct{}

func (p *ReadMeOSSParser) Parse(path string) (*model.ParsedReport, error) {
	return ParseReadMeOSS(path)
}

func (p *ReadMeOSSParser) Format() string {
	return "readmeoss"
}
