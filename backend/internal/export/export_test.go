package export

import (
	"encoding/json"
	"testing"

	"github.com/fossology/report-aggregator/internal/model"
)

func makeMergeResult() *model.MergeResult {
	return &model.MergeResult{
		Components: []model.Component{
			{
				Name:        "lodash",
				Version:     "4.17.21",
				PURL:        "pkg:npm/lodash@4.17.21",
				Supplier:    "Lodash Team",
				Description: "A modern JavaScript utility library",
				DownloadURL: "https://registry.npmjs.org/lodash/-/lodash-4.17.21.tgz",
				Licenses: []model.License{
					{ID: "MIT", Name: "MIT License"},
				},
				Copyrights: []model.Copyright{
					{Text: "Copyright JS Foundation"},
				},
				Hashes: []model.Hash{
					{Algorithm: "SHA-256", Value: "abc123def456"},
				},
			},
			{
				Name:    "express",
				Version: "4.18.2",
				Licenses: []model.License{
					{ID: "MIT"},
				},
			},
		},
		SourceReports: []string{"report1.spdx.json", "report2.cdx.json"},
		MergedAt:      "2024-01-15T10:00:00Z",
	}
}

func TestToSPDX(t *testing.T) {
	result := makeMergeResult()
	data, err := ToSPDX(result)
	if err != nil {
		t.Fatalf("ToSPDX() error: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("ToSPDX() returned empty data")
	}

	// Verify it's valid JSON
	var doc map[string]interface{}
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("ToSPDX() output is not valid JSON: %v", err)
	}

	// Check required SPDX fields
	if doc["spdxVersion"] != "SPDX-2.3" {
		t.Errorf("spdxVersion = %v, want %q", doc["spdxVersion"], "SPDX-2.3")
	}
	if doc["dataLicense"] != "CC0-1.0" {
		t.Errorf("dataLicense = %v, want %q", doc["dataLicense"], "CC0-1.0")
	}

	// Check packages are present
	packages, ok := doc["packages"].([]interface{})
	if !ok {
		t.Fatal("packages field missing or wrong type")
	}
	if len(packages) != 2 {
		t.Errorf("expected 2 packages, got %d", len(packages))
	}
}

func TestToCycloneDX(t *testing.T) {
	result := makeMergeResult()
	data, err := ToCycloneDX(result)
	if err != nil {
		t.Fatalf("ToCycloneDX() error: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("ToCycloneDX() returned empty data")
	}

	// Verify it's valid JSON
	var doc map[string]interface{}
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("ToCycloneDX() output is not valid JSON: %v", err)
	}

	// Check required CycloneDX fields
	if doc["bomFormat"] != "CycloneDX" {
		t.Errorf("bomFormat = %v, want %q", doc["bomFormat"], "CycloneDX")
	}

	// Check components are present
	components, ok := doc["components"].([]interface{})
	if !ok {
		t.Fatal("components field missing or wrong type")
	}
	if len(components) != 2 {
		t.Errorf("expected 2 components, got %d", len(components))
	}
}

func TestToSPDXEmptyResult(t *testing.T) {
	result := &model.MergeResult{}
	data, err := ToSPDX(result)
	if err != nil {
		t.Fatalf("ToSPDX() with empty result error: %v", err)
	}

	var doc map[string]interface{}
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
}

func TestToCycloneDXEmptyResult(t *testing.T) {
	result := &model.MergeResult{}
	data, err := ToCycloneDX(result)
	if err != nil {
		t.Fatalf("ToCycloneDX() with empty result error: %v", err)
	}

	var doc map[string]interface{}
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
}
