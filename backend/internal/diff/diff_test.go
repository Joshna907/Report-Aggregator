package diff

import (
	"testing"

	"github.com/fossology/report-aggregator/internal/model"
)

func makeComponent(name, version, purl, supplier, license string) model.Component {
	comp := model.Component{
		Name:     name,
		Version:  version,
		PURL:     purl,
		Supplier: supplier,
	}
	if license != "" {
		comp.Licenses = []model.License{{ID: license}}
	}
	return comp
}

func TestDetectChanges(t *testing.T) {
	oldReport := &model.ParsedReport{
		Components: []model.Component{
			makeComponent("lodash", "4.17.21", "pkg:npm/lodash@4.17.21", "Lodash Team", "MIT"),
			makeComponent("express", "4.18.2", "", "", "MIT"),
			makeComponent("react", "18.2.0", "", "", "MIT"),
		},
	}

	newReport := &model.ParsedReport{
		Components: []model.Component{
			// Modified: license changed, supplier added
			makeComponent("lodash", "4.17.21", "pkg:npm/lodash@4.17.21", "New Team", "Apache-2.0"),
			// Unchanged
			makeComponent("express", "4.18.2", "", "", "MIT"),
			// Added
			makeComponent("axios", "1.6.0", "", "Axios Team", "MIT"),
			// react is removed
		},
	}

	diff := DetectChanges(oldReport, newReport)

	if len(diff.Added) != 1 || diff.Added[0].Name != "axios" {
		t.Errorf("expected 1 added component ('axios'), got %d", len(diff.Added))
	}

	if len(diff.Removed) != 1 || diff.Removed[0].Name != "react" {
		t.Errorf("expected 1 removed component ('react'), got %d", len(diff.Removed))
	}

	if len(diff.Modified) != 1 || diff.Modified[0].Name != "lodash" {
		t.Fatalf("expected 1 modified component ('lodash'), got %d", len(diff.Modified))
	}

	mod := diff.Modified[0]
	if len(mod.Changes) != 2 {
		t.Fatalf("expected 2 changes for lodash, got %d", len(mod.Changes))
	}

	// Check specific changes
	var licChange, supChange bool
	for _, c := range mod.Changes {
		if c.Field == "license" && c.OldValue == "MIT" && c.NewValue == "Apache-2.0" {
			licChange = true
		}
		if c.Field == "supplier" && c.OldValue == "Lodash Team" && c.NewValue == "New Team" {
			supChange = true
		}
	}

	if !licChange {
		t.Error("failed to detect license change correctly")
	}
	if !supChange {
		t.Error("failed to detect supplier change correctly")
	}
}

func TestLicensesAsString(t *testing.T) {
	lics := []model.License{
		{ID: "MIT"},
		{ID: "Apache-2.0"},
	}
	
	// Should sort them automatically
	result := licensesAsString(lics)
	if result != "Apache-2.0, MIT" {
		t.Errorf("expected 'Apache-2.0, MIT', got %q", result)
	}
}
