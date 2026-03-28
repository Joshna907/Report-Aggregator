package parser

import (
	"path/filepath"
	"runtime"
	"testing"
)

// testdataDir returns the absolute path to the testdata directory
func testdataDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "..", "..", "testdata")
}

func TestDetectFormat(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"report.spdx.json", "spdx"},
		{"report.cdx.json", "cyclonedx"},
		{"my-spdx-report.json", "spdx"},
		{"cyclonedx-bom.json", "cyclonedx"},
		{"report.cli.xml", "clixml"},
		{"readmeoss_report.txt", "readmeoss"},
		{"readme_oss_data.txt", "readmeoss"},
		{"report.spdx", "spdx"},
		{"report.xml", "cyclonedx"},
		{"unknown.csv", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := DetectFormat(tt.filename)
			if got != tt.expected {
				t.Errorf("DetectFormat(%q) = %q, want %q", tt.filename, got, tt.expected)
			}
		})
	}
}

func TestGetParser(t *testing.T) {
	formats := []string{"spdx", "cyclonedx", "clixml", "readmeoss"}
	for _, f := range formats {
		t.Run(f, func(t *testing.T) {
			p, err := GetParser(f)
			if err != nil {
				t.Fatalf("GetParser(%q) returned error: %v", f, err)
			}
			if p.Format() != f {
				t.Errorf("parser.Format() = %q, want %q", p.Format(), f)
			}
		})
	}

	// Test unsupported format
	_, err := GetParser("unknown")
	if err == nil {
		t.Error("GetParser(\"unknown\") should return an error")
	}
}

func TestSPDXParser(t *testing.T) {
	path := filepath.Join(testdataDir(), "sample.spdx.json")
	parser := &SPDXParser{}

	report, err := parser.Parse(path)
	if err != nil {
		t.Fatalf("SPDXParser.Parse() error: %v", err)
	}

	if report.Format != "spdx" {
		t.Errorf("report.Format = %q, want %q", report.Format, "spdx")
	}

	if len(report.Components) != 3 {
		t.Fatalf("expected 3 components, got %d", len(report.Components))
	}

	// Verify first component (lodash)
	lodash := report.Components[0]
	if lodash.Name != "lodash" {
		t.Errorf("first component name = %q, want %q", lodash.Name, "lodash")
	}
	if lodash.Version != "4.17.21" {
		t.Errorf("lodash version = %q, want %q", lodash.Version, "4.17.21")
	}
	if lodash.PURL != "pkg:npm/lodash@4.17.21" {
		t.Errorf("lodash PURL = %q, want %q", lodash.PURL, "pkg:npm/lodash@4.17.21")
	}
	if len(lodash.Licenses) == 0 {
		t.Fatal("lodash should have at least 1 license")
	}
	if lodash.Licenses[0].ID != "MIT" {
		t.Errorf("lodash license = %q, want %q", lodash.Licenses[0].ID, "MIT")
	}
	if len(lodash.Copyrights) == 0 {
		t.Fatal("lodash should have at least 1 copyright")
	}
	if len(lodash.Hashes) == 0 {
		t.Fatal("lodash should have at least 1 hash")
	}

	// Verify provenance is set
	if len(lodash.Provenance) == 0 || lodash.Provenance[0].Format != "spdx" {
		format := "none"
		if len(lodash.Provenance) > 0 {
			format = lodash.Provenance[0].Format
		}
		t.Errorf("provenance format = %q, want %q", format, "spdx")
	}
}

func TestCycloneDXParser(t *testing.T) {
	path := filepath.Join(testdataDir(), "sample.cdx.json")
	parser := &CycloneDXParser{}

	report, err := parser.Parse(path)
	if err != nil {
		t.Fatalf("CycloneDXParser.Parse() error: %v", err)
	}

	if report.Format != "cyclonedx" {
		t.Errorf("report.Format = %q, want %q", report.Format, "cyclonedx")
	}

	if len(report.Components) != 3 {
		t.Fatalf("expected 3 components, got %d", len(report.Components))
	}

	// Verify lodash
	lodash := report.Components[0]
	if lodash.Name != "lodash" {
		t.Errorf("first component = %q, want %q", lodash.Name, "lodash")
	}
	if lodash.PURL != "pkg:npm/lodash@4.17.21" {
		t.Errorf("lodash PURL = %q, want %q", lodash.PURL, "pkg:npm/lodash@4.17.21")
	}
	if lodash.Supplier != "Lodash Team" {
		t.Errorf("lodash supplier = %q, want %q", lodash.Supplier, "Lodash Team")
	}

	// Verify axios (unique to CycloneDX)
	axios := report.Components[1]
	if axios.Name != "axios" {
		t.Errorf("second component = %q, want %q", axios.Name, "axios")
	}
}

func TestCLIXMLParser(t *testing.T) {
	path := filepath.Join(testdataDir(), "sample.cli.xml")
	parser := &CLIXMLParser{}

	report, err := parser.Parse(path)
	if err != nil {
		t.Fatalf("CLIXMLParser.Parse() error: %v", err)
	}

	if report.Format != "clixml" {
		t.Errorf("report.Format = %q, want %q", report.Format, "clixml")
	}

	if len(report.Components) != 2 {
		t.Fatalf("expected 2 components, got %d", len(report.Components))
	}

	// Verify zlib
	zlib := report.Components[0]
	if zlib.Name != "zlib" {
		t.Errorf("first component = %q, want %q", zlib.Name, "zlib")
	}
	if zlib.Version != "1.2.13" {
		t.Errorf("zlib version = %q, want %q", zlib.Version, "1.2.13")
	}
	if len(zlib.Licenses) == 0 {
		t.Fatal("zlib should have at least 1 license")
	}
	if zlib.Licenses[0].ID != "Zlib" {
		t.Errorf("zlib license SPDX ID = %q, want %q", zlib.Licenses[0].ID, "Zlib")
	}
	if len(zlib.Copyrights) == 0 {
		t.Fatal("zlib should have at least 1 copyright")
	}
}

func TestReadMeOSSParser(t *testing.T) {
	path := filepath.Join(testdataDir(), "sample.readmeoss.txt")
	parser := &ReadMeOSSParser{}

	report, err := parser.Parse(path)
	if err != nil {
		t.Fatalf("ReadMeOSSParser.Parse() error: %v", err)
	}

	if report.Format != "readmeoss" {
		t.Errorf("report.Format = %q, want %q", report.Format, "readmeoss")
	}

	if len(report.Components) != 2 {
		t.Fatalf("expected 2 components, got %d", len(report.Components))
	}

	// Verify curl
	curl := report.Components[0]
	if curl.Name != "curl" {
		t.Errorf("first component = %q, want %q", curl.Name, "curl")
	}
	if curl.Version != "8.4.0" {
		t.Errorf("curl version = %q, want %q", curl.Version, "8.4.0")
	}

	// Verify libpng
	libpng := report.Components[1]
	if libpng.Name != "libpng" {
		t.Errorf("second component = %q, want %q", libpng.Name, "libpng")
	}
}

func TestDetectAndParse(t *testing.T) {
	path := filepath.Join(testdataDir(), "sample.spdx.json")
	report, err := DetectAndParse(path)
	if err != nil {
		t.Fatalf("DetectAndParse() error: %v", err)
	}
	if report.Format != "spdx" {
		t.Errorf("report.Format = %q, want %q", report.Format, "spdx")
	}
	if len(report.Components) == 0 {
		t.Error("expected at least 1 component")
	}
}

func TestParseNonExistentFile(t *testing.T) {
	path := filepath.Join(testdataDir(), "nonexistent.spdx.json")
	_, err := DetectAndParse(path)
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}
