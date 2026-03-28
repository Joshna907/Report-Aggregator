package model

import "time"

// Provenance tracks where a piece of data came from
type Provenance struct {
	ReportFile string `json:"reportFile"`
	Format     string `json:"format"` // "spdx", "cyclonedx", "clixml", "readmeoss"
	ParsedAt   string `json:"parsedAt"`
}

// License represents a license with provenance
type License struct {
	ID         string       `json:"id"`         // SPDX identifier e.g. "MIT", "Apache-2.0"
	Name       string       `json:"name"`       // Full name
	Text       string       `json:"text,omitempty"`
	Provenance []Provenance `json:"provenance"`
}

// Copyright represents a copyright statement with provenance
type Copyright struct {
	Text       string       `json:"text"`
	Provenance []Provenance `json:"provenance"`
}

// Hash represents a file hash
type Hash struct {
	Algorithm string `json:"algorithm"` // "SHA-1", "SHA-256", "MD5"
	Value     string `json:"value"`
}

// Component represents a single software component in the unified model
type Component struct {
	ID          int64        `json:"id,omitempty"`       // Database ID
	ReportID    int64        `json:"reportId,omitempty"` // Link to source report
	Name        string       `json:"name"`
	Version     string       `json:"version"`
	PURL        string       `json:"purl,omitempty"`  // Package URL — best identifier
	Supplier    string       `json:"supplier,omitempty"`
	Licenses    []License    `json:"licenses"`
	Copyrights  []Copyright  `json:"copyrights,omitempty"`
	Hashes      []Hash       `json:"hashes,omitempty"`
	Description string       `json:"description,omitempty"`
	DownloadURL string      `json:"downloadUrl,omitempty"`
	Provenance  []Provenance `json:"provenance"` // Sources this component came from
}

// Relationship represents a dependency or structural link between components
type Relationship struct {
	Source     string       `json:"source"`     // Name of the source component
	Target     string       `json:"target"`     // Name of the target component
	Type       string       `json:"type"`       // e.g. "DEPENDS_ON", "CONTAINS"
	Provenance []Provenance `json:"provenance"` // Where this relationship was found
}

// Conflict represents a disagreement between two reports about the same component
type Conflict struct {
	ID            int64     `json:"id,omitempty"`
	ReportID      int64     `json:"reportId,omitempty"`
	ComponentName string    `json:"componentName"`
	ComponentVer  string    `json:"componentVersion"`
	Field         string    `json:"field"` // "license", "copyright", "supplier"
	ValueA        string    `json:"valueA"`
	SourceA       string    `json:"sourceA"` // report file name
	ValueB        string    `json:"valueB"`
	SourceB       string    `json:"sourceB"`
	Resolved      bool      `json:"resolved"`
	Resolution    string    `json:"resolution,omitempty"`
	ResolvedBy    string    `json:"resolvedBy,omitempty"`
	ResolvedAt    time.Time `json:"resolvedAt,omitempty"`
}

// MergeResult holds the output of merging multiple reports
type MergeResult struct {
	ID            int64          `json:"id,omitempty"`
	Components    []Component    `json:"components"`
	Relationships []Relationship `json:"relationships,omitempty"`
	Conflicts     []Conflict     `json:"conflicts"`
	SourceReports []string    `json:"sourceReports"` // list of input file names
	MergedAt      string      `json:"mergedAt"`
	Summary       Summary     `json:"summary"`
}

// Summary provides high-level stats about the merged report
type Summary struct {
	TotalComponents   int            `json:"totalComponents"`
	TotalLicenses     int            `json:"totalLicenses"`
	TotalConflicts    int            `json:"totalConflicts"`
	SourceReportsCount int           `json:"sourceReportsCount"`
	LicenseBreakdown  map[string]int `json:"licenseBreakdown"`  // license ID -> count
	SourceBreakdown   map[string]int `json:"sourceBreakdown"`   // report file -> component count
}

// ParsedReport represents a single parsed report before merging
type ParsedReport struct {
	ID            int64          `json:"id,omitempty"`
	FileName      string         `json:"fileName"`
	Format        string         `json:"format"`
	Components    []Component    `json:"components"`
	Relationships []Relationship `json:"relationships,omitempty"`
	ParsedAt      time.Time      `json:"parsedAt"`
}

// ValidationError represents an issue found during report validation
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Line    int    `json:"line,omitempty"`
}

// ValidationResult holds the result of validating a report
type ValidationResult struct {
	Valid    bool              `json:"valid"`
	Errors   []ValidationError `json:"errors,omitempty"`
	Warnings []ValidationError `json:"warnings,omitempty"`
}
