package validate

import (
	"testing"

	"github.com/fossology/report-aggregator/internal/model"
)

func TestValidateEmptyReport(t *testing.T) {
	report := &model.ParsedReport{
		FileName: "test.spdx.json",
		Format:   "spdx",
	}

	result := Validate(report)
	if result.Valid {
		t.Error("empty report should not be valid")
	}
	if len(result.Errors) == 0 {
		t.Error("expected validation errors for empty report")
	}
}

func TestValidateComponentWithoutName(t *testing.T) {
	report := &model.ParsedReport{
		FileName: "test.spdx.json",
		Format:   "spdx",
		Components: []model.Component{
			{Name: "", Version: "1.0.0"},
		},
	}

	result := Validate(report)
	if result.Valid {
		t.Error("component without name should make report invalid")
	}
}

func TestValidateComponentWithoutVersion(t *testing.T) {
	report := &model.ParsedReport{
		FileName: "test.spdx.json",
		Format:   "spdx",
		Components: []model.Component{
			{Name: "lodash", Version: ""},
		},
	}

	result := Validate(report)
	// Missing version should be a warning, not an error
	if !result.Valid {
		t.Error("missing version should be a warning, not invalidate the report")
	}
	if len(result.Warnings) == 0 {
		t.Error("expected warning for missing version")
	}
}

func TestValidateComponentWithoutLicense(t *testing.T) {
	report := &model.ParsedReport{
		FileName: "test.spdx.json",
		Format:   "spdx",
		Components: []model.Component{
			{Name: "lodash", Version: "4.17.21"},
		},
	}

	result := Validate(report)
	if !result.Valid {
		t.Error("missing license should be a warning, not invalidate")
	}
	if len(result.Warnings) == 0 {
		t.Error("expected warning for missing license")
	}
}

func TestValidateEmptyLicenseIDAndName(t *testing.T) {
	report := &model.ParsedReport{
		FileName: "test.spdx.json",
		Format:   "spdx",
		Components: []model.Component{
			{
				Name:    "lodash",
				Version: "4.17.21",
				Licenses: []model.License{
					{ID: "", Name: ""},
				},
			},
		},
	}

	result := Validate(report)
	if len(result.Warnings) == 0 {
		t.Error("expected warning for license with no ID or name")
	}
}

func TestValidateValidReport(t *testing.T) {
	report := &model.ParsedReport{
		FileName: "test.spdx.json",
		Format:   "spdx",
		Components: []model.Component{
			{
				Name:    "lodash",
				Version: "4.17.21",
				Licenses: []model.License{
					{ID: "MIT", Name: "MIT License"},
				},
			},
		},
	}

	result := Validate(report)
	if !result.Valid {
		t.Error("valid report should pass validation")
	}
	if len(result.Errors) != 0 {
		t.Errorf("expected 0 errors, got %d", len(result.Errors))
	}
}
