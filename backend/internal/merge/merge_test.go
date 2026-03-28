package merge

import (
	"testing"

	"github.com/fossology/report-aggregator/internal/model"
)

func makeComponent(name, version, purl, license, source string) model.Component {
	return model.Component{
		Name:    name,
		Version: version,
		PURL:    purl,
		Licenses: []model.License{
			{ID: license, Provenance: []model.Provenance{{ReportFile: source, Format: "test"}}},
		},
		Provenance: []model.Provenance{{ReportFile: source, Format: "test"}},
	}
}

func makeReport(fileName string, components ...model.Component) *model.ParsedReport {
	return &model.ParsedReport{
		FileName:   fileName,
		Format:     "test",
		Components: components,
	}
}

func TestMergeEmptyReports(t *testing.T) {
	_, err := Merge(nil)
	if err == nil {
		t.Error("expected error for nil reports")
	}

	_, err = Merge([]*model.ParsedReport{})
	if err == nil {
		t.Error("expected error for empty reports")
	}
}

func TestMergeSingleReport(t *testing.T) {
	report := makeReport("report1.spdx",
		makeComponent("lodash", "4.17.21", "pkg:npm/lodash@4.17.21", "MIT", "report1.spdx"),
		makeComponent("express", "4.18.2", "", "MIT", "report1.spdx"),
	)

	result, err := Merge([]*model.ParsedReport{report})
	if err != nil {
		t.Fatalf("Merge() error: %v", err)
	}

	if len(result.Components) != 2 {
		t.Errorf("expected 2 components, got %d", len(result.Components))
	}
	if len(result.Conflicts) != 0 {
		t.Errorf("expected 0 conflicts, got %d", len(result.Conflicts))
	}
	if len(result.SourceReports) != 1 {
		t.Errorf("expected 1 source report, got %d", len(result.SourceReports))
	}
}

func TestMergeSmartMatchingByPURL(t *testing.T) {
	report1 := makeReport("report1.spdx",
		makeComponent("lodash", "4.17.21", "pkg:npm/lodash@4.17.21", "MIT", "report1.spdx"),
	)
	report2 := makeReport("report2.cdx",
		makeComponent("Lodash", "v4.17.21", "pkg:npm/lodash@4.17.21", "MIT", "report2.cdx"),
	)

	result, err := Merge([]*model.ParsedReport{report1, report2})
	if err != nil {
		t.Fatalf("Merge() error: %v", err)
	}

	// Should match by PURL — result should have 1 component, not 2
	if len(result.Components) != 1 {
		t.Errorf("expected 1 component (PURL match), got %d", len(result.Components))
	}
}

func TestMergeSmartMatchingByNameVersion(t *testing.T) {
	report1 := makeReport("report1.spdx",
		makeComponent("express", "4.18.2", "", "MIT", "report1.spdx"),
	)
	report2 := makeReport("report2.cdx",
		makeComponent("express", "4.18.2", "", "MIT", "report2.cdx"),
	)

	result, err := Merge([]*model.ParsedReport{report1, report2})
	if err != nil {
		t.Fatalf("Merge() error: %v", err)
	}

	if len(result.Components) != 1 {
		t.Errorf("expected 1 component (name+version match), got %d", len(result.Components))
	}
}

func TestMergeVersionNormalization(t *testing.T) {
	report1 := makeReport("report1.spdx",
		makeComponent("react", "18.2.0", "", "MIT", "report1.spdx"),
	)
	report2 := makeReport("report2.cdx",
		makeComponent("react", "v18.2.0", "", "MIT", "report2.cdx"),
	)

	result, err := Merge([]*model.ParsedReport{report1, report2})
	if err != nil {
		t.Fatalf("Merge() error: %v", err)
	}

	// "18.2.0" and "v18.2.0" should match after normalization
	if len(result.Components) != 1 {
		t.Errorf("expected 1 component (version normalization), got %d", len(result.Components))
	}
}

func TestMergeSmartMatchingByHash(t *testing.T) {
	comp1 := makeComponent("unknown-pkg", "", "", "MIT", "report1.spdx")
	comp1.Hashes = []model.Hash{{Algorithm: "SHA-256", Value: "abc123"}}

	comp2 := makeComponent("real-pkg", "1.0.0", "", "MIT", "report2.cdx")
	comp2.Hashes = []model.Hash{{Algorithm: "SHA-256", Value: "abc123"}}

	report1 := makeReport("report1.spdx", comp1)
	report2 := makeReport("report2.cdx", comp2)

	result, err := Merge([]*model.ParsedReport{report1, report2})
	if err != nil {
		t.Fatalf("Merge() error: %v", err)
	}

	// Should match by hash
	if len(result.Components) != 1 {
		t.Errorf("expected 1 component (hash match), got %d", len(result.Components))
	}
}

func TestMergeConflictDetection(t *testing.T) {
	report1 := makeReport("report1.spdx",
		makeComponent("react", "18.2.0", "pkg:npm/react@18.2.0", "MIT", "report1.spdx"),
	)
	report2 := makeReport("report2.cdx",
		makeComponent("react", "18.2.0", "pkg:npm/react@18.2.0", "Apache-2.0", "report2.cdx"),
	)

	result, err := Merge([]*model.ParsedReport{report1, report2})
	if err != nil {
		t.Fatalf("Merge() error: %v", err)
	}

	if len(result.Components) != 1 {
		t.Errorf("expected 1 component, got %d", len(result.Components))
	}

	// Should detect a license conflict
	if len(result.Conflicts) == 0 {
		t.Fatal("expected at least 1 conflict for different licenses")
	}

	conflict := result.Conflicts[0]
	if conflict.Field != "license" {
		t.Errorf("conflict field = %q, want %q", conflict.Field, "license")
	}
	if conflict.ComponentName != "react" {
		t.Errorf("conflict component = %q, want %q", conflict.ComponentName, "react")
	}
}

func TestMergeUnionLicenses(t *testing.T) {
	comp1 := makeComponent("lodash", "4.17.21", "pkg:npm/lodash@4.17.21", "MIT", "report1.spdx")
	comp2 := makeComponent("lodash", "4.17.21", "pkg:npm/lodash@4.17.21", "MIT", "report2.cdx")
	// Add a second license to comp2
	comp2.Licenses = append(comp2.Licenses, model.License{ID: "ISC", Provenance: []model.Provenance{{ReportFile: "report2.cdx"}}})

	report1 := makeReport("report1.spdx", comp1)
	report2 := makeReport("report2.cdx", comp2)

	result, err := Merge([]*model.ParsedReport{report1, report2})
	if err != nil {
		t.Fatalf("Merge() error: %v", err)
	}

	// Should have both MIT and ISC licenses after union merge
	if len(result.Components[0].Licenses) < 2 {
		t.Errorf("expected at least 2 licenses after union merge, got %d", len(result.Components[0].Licenses))
	}
}

func TestMergeUniqueComponents(t *testing.T) {
	report1 := makeReport("report1.spdx",
		makeComponent("lodash", "4.17.21", "", "MIT", "report1.spdx"),
	)
	report2 := makeReport("report2.cdx",
		makeComponent("axios", "1.6.0", "", "MIT", "report2.cdx"),
	)

	result, err := Merge([]*model.ParsedReport{report1, report2})
	if err != nil {
		t.Fatalf("Merge() error: %v", err)
	}

	// Unique components should both appear
	if len(result.Components) != 2 {
		t.Errorf("expected 2 components, got %d", len(result.Components))
	}
}

func TestMergeSummary(t *testing.T) {
	report1 := makeReport("report1.spdx",
		makeComponent("lodash", "4.17.21", "", "MIT", "report1.spdx"),
		makeComponent("express", "4.18.2", "", "Apache-2.0", "report1.spdx"),
	)

	result, err := Merge([]*model.ParsedReport{report1})
	if err != nil {
		t.Fatalf("Merge() error: %v", err)
	}

	if result.Summary.TotalComponents != 2 {
		t.Errorf("summary.TotalComponents = %d, want 2", result.Summary.TotalComponents)
	}
	if result.Summary.TotalLicenses != 2 {
		t.Errorf("summary.TotalLicenses = %d, want 2", result.Summary.TotalLicenses)
	}
	if result.Summary.LicenseBreakdown["MIT"] != 1 {
		t.Errorf("MIT count = %d, want 1", result.Summary.LicenseBreakdown["MIT"])
	}
	if result.Summary.LicenseBreakdown["Apache-2.0"] != 1 {
		t.Errorf("Apache-2.0 count = %d, want 1", result.Summary.LicenseBreakdown["Apache-2.0"])
	}
}

func TestMergeFillsEmptyFields(t *testing.T) {
	comp1 := makeComponent("lodash", "4.17.21", "", "MIT", "report1.spdx")
	comp2 := makeComponent("lodash", "4.17.21", "pkg:npm/lodash@4.17.21", "MIT", "report2.cdx")
	comp2.Supplier = "Lodash Team"
	comp2.Description = "A modern JavaScript utility library"

	report1 := makeReport("report1.spdx", comp1)
	report2 := makeReport("report2.cdx", comp2)

	result, err := Merge([]*model.ParsedReport{report1, report2})
	if err != nil {
		t.Fatalf("Merge() error: %v", err)
	}

	merged := result.Components[0]
	if merged.PURL != "pkg:npm/lodash@4.17.21" {
		t.Errorf("PURL not filled from second report: %q", merged.PURL)
	}
	if merged.Supplier != "Lodash Team" {
		t.Errorf("Supplier not filled from second report: %q", merged.Supplier)
	}
	if merged.Description != "A modern JavaScript utility library" {
		t.Errorf("Description not filled from second report: %q", merged.Description)
	}
}
