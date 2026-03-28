package parser

import (
	"testing"
)

func TestSPDX3Parser(t *testing.T) {
	p := &SPDX3Parser{}
	report, err := p.Parse("../../testdata/sample.spdx3.json")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if report.Format != "spdx3" {
		t.Errorf("expected format spdx3, got %s", report.Format)
	}

	if len(report.Components) != 2 {
		t.Fatalf("expected 2 components, got %d", len(report.Components))
	}

	// Check first component (Package type)
	c1 := report.Components[0]
	if c1.Name != "react" || c1.Version != "18.2.0" {
		t.Errorf("unexpected component 1: %s@%s", c1.Name, c1.Version)
	}
	if c1.PURL != "pkg:npm/react@18.2.0" {
		t.Errorf("expected PURL pkg:npm/react@18.2.0, got %s", c1.PURL)
	}
	if len(c1.Licenses) == 0 || c1.Licenses[0].ID != "MIT" {
		t.Errorf("expected license MIT, got %v", c1.Licenses)
	}

	// Check second component (Software type)
	c2 := report.Components[1]
	if c2.Name != "lodash" || c2.Version != "4.17.21" {
		t.Errorf("unexpected component 2: %s@%s", c2.Name, c2.Version)
	}
	if len(c2.Licenses) == 0 || c2.Licenses[0].ID != "MIT" {
		t.Errorf("expected license MIT for lodash, got %v", c2.Licenses)
	}
}
