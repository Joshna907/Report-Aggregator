package validate

import (
	"fmt"

	"github.com/fossology/report-aggregator/internal/model"
)

// Validate checks a parsed report for completeness and correctness
func Validate(report *model.ParsedReport) *model.ValidationResult {
	result := &model.ValidationResult{
		Valid: true,
	}

	if len(report.Components) == 0 {
		result.Errors = append(result.Errors, model.ValidationError{
			Field:   "components",
			Message: "Report contains no components",
		})
		result.Valid = false
		return result
	}

	for i, comp := range report.Components {
		// Check required fields
		if comp.Name == "" {
			result.Errors = append(result.Errors, model.ValidationError{
				Field:   "name",
				Message: "Component at index " + itoa(i) + " has no name",
			})
			result.Valid = false
		}

		if comp.Version == "" {
			result.Warnings = append(result.Warnings, model.ValidationError{
				Field:   "version",
				Message: "Component '" + comp.Name + "' has no version",
			})
		}

		if len(comp.Licenses) == 0 {
			result.Warnings = append(result.Warnings, model.ValidationError{
				Field:   "licenses",
				Message: "Component '" + comp.Name + "' has no license information",
			})
		}

		// Check for empty license IDs
		for _, lic := range comp.Licenses {
			if lic.ID == "" && lic.Name == "" {
				result.Warnings = append(result.Warnings, model.ValidationError{
					Field:   "licenses",
					Message: "Component '" + comp.Name + "' has a license with no ID or name",
				})
			}
		}
	}

	return result
}

func itoa(i int) string {
	return fmt.Sprintf("%d", i)
}
