package parser

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/fossology/report-aggregator/internal/model"
)

// Parser interface — all format parsers implement this
type Parser interface {
	Parse(filePath string) (*model.ParsedReport, error)
	Format() string
}

// DetectAndParse auto-detects the file format and parses it
func DetectAndParse(filePath string) (*model.ParsedReport, error) {
	format := DetectFormat(filePath)
	parser, err := GetParser(format)
	if err != nil {
		return nil, err
	}
	return parser.Parse(filePath)
}

// DetectFormat determines the report format based on file extension and name
func DetectFormat(filePath string) string {
	name := strings.ToLower(filepath.Base(filePath))
	ext := strings.ToLower(filepath.Ext(filePath))

	// Check filename patterns first
	if strings.Contains(name, "cyclonedx") || strings.Contains(name, "bom") {
		return "cyclonedx"
	}
	if strings.Contains(name, "spdx3") {
		return "spdx3"
	}
	if strings.Contains(name, "spdx") {
		return "spdx"
	}
	if strings.Contains(name, "cli") && ext == ".xml" {
		return "clixml"
	}
	if strings.Contains(name, "readmeoss") || strings.Contains(name, "readme_oss") {
		return "readmeoss"
	}

	// Check for double extensions (e.g. .cdx.json, .spdx.json)
	if strings.HasSuffix(name, ".cdx.json") || strings.HasSuffix(name, ".cdx.xml") {
		return "cyclonedx"
	}
	if strings.HasSuffix(name, ".spdx3.json") {
		return "spdx3"
	}
	if strings.HasSuffix(name, ".spdx.json") || strings.HasSuffix(name, ".spdx.rdf") {
		return "spdx"
	}

	// Fall back to single extension
	switch ext {
	case ".spdx", ".rdf", ".tv":
		return "spdx"
	case ".xml":
		// Could be CycloneDX or CLIXML — default to cyclonedx
		return "cyclonedx"
	case ".json":
		// Could be SPDX or CycloneDX — default to spdx
		return "spdx"
	case ".txt":
		return "readmeoss"
	default:
		return "unknown"
	}
}

// GetParser returns the appropriate parser for a format
func GetParser(format string) (Parser, error) {
	switch format {
	case "spdx":
		return &SPDXParser{}, nil
	case "spdx3":
		return &SPDX3Parser{}, nil
	case "cyclonedx":
		return &CycloneDXParser{}, nil
	case "clixml":
		return &CLIXMLParser{}, nil
	case "readmeoss":
		return &ReadMeOSSParser{}, nil
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}
