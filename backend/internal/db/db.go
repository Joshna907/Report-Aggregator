package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "modernc.org/sqlite"

	"github.com/fossology/report-aggregator/internal/model"
)

type DB struct {
	Conn *sql.DB
}

// InitDB initializes the SQLite database
func InitDB(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db := &DB{Conn: conn}
	if err := db.createSchema(); err != nil {
		return nil, err
	}

	return db, nil
}

func (db *DB) createSchema() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS reports (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			file_name TEXT,
			format TEXT,
			parsed_at DATETIME,
			summary TEXT     -- JSON serialized Summary
		)`,
		`CREATE TABLE IF NOT EXISTS components (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			report_id INTEGER,
			name TEXT,
			version TEXT,
			purl TEXT,
			supplier TEXT,
			description TEXT,
			download_url TEXT,
			licenses TEXT,   -- JSON serialized
			copyrights TEXT, -- JSON serialized
			hashes TEXT,     -- JSON serialized
			provenance TEXT, -- JSON serialized
			FOREIGN KEY(report_id) REFERENCES reports(id)
		)`,
		`CREATE TABLE IF NOT EXISTS relationships (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			report_id INTEGER,
			source TEXT,
			target TEXT,
			type TEXT,
			provenance TEXT, -- JSON serialized
			FOREIGN KEY(report_id) REFERENCES reports(id)
		)`,
		`CREATE TABLE IF NOT EXISTS conflicts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			report_id INTEGER,
			component_name TEXT,
			component_version TEXT,
			field TEXT,
			value_a TEXT,
			source_a TEXT,
			value_b TEXT,
			source_b TEXT,
			resolved BOOLEAN,
			resolution TEXT,
			resolved_by TEXT,
			resolved_at DATETIME,
			FOREIGN KEY(report_id) REFERENCES reports(id)
		)`,
		`CREATE TABLE IF NOT EXISTS changelog (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			report_id INTEGER,
			component_name TEXT,
			field TEXT,
			old_value TEXT,
			new_value TEXT,
			changed_by TEXT,
			reason TEXT,
			timestamp DATETIME
		)`,
	}

	for _, q := range queries {
		if _, err := db.Conn.Exec(q); err != nil {
			return fmt.Errorf("failed to create schema: %w", err)
		}
	}

	// ---------------------------------------------------------
	// MIGRATIONS
	// ---------------------------------------------------------
	// Try to add 'summary' column to 'reports' table if it doesn't exist
	_, _ = db.Conn.Exec(`ALTER TABLE reports ADD COLUMN summary TEXT`)

	return nil
}

// SaveReport saves a parsed report and its components/relationships to the DB
func (db *DB) SaveReport(report *model.ParsedReport) error {
	tx, err := db.Conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.Exec(`INSERT INTO reports (file_name, format, parsed_at) VALUES (?, ?, ?)`,
		report.FileName, report.Format, report.ParsedAt)
	if err != nil {
		return err
	}

	reportID, _ := res.LastInsertId()
	report.ID = reportID

	for i := range report.Components {
		comp := &report.Components[i]
		comp.ReportID = reportID

		licJson, _ := json.Marshal(comp.Licenses)
		copyJson, _ := json.Marshal(comp.Copyrights)
		hashJson, _ := json.Marshal(comp.Hashes)
		provJson, _ := json.Marshal(comp.Provenance)

		resComp, err := tx.Exec(`INSERT INTO components (report_id, name, version, purl, supplier, description, download_url, licenses, copyrights, hashes, provenance) 
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			reportID, comp.Name, comp.Version, comp.PURL, comp.Supplier, comp.Description, comp.DownloadURL,
			string(licJson), string(copyJson), string(hashJson), string(provJson))
		if err != nil {
			return err
		}
		comp.ID, _ = resComp.LastInsertId()
	}

	for _, rel := range report.Relationships {
		relProvJson, _ := json.Marshal(rel.Provenance)
		_, err := tx.Exec(`INSERT INTO relationships (report_id, source, target, type, provenance) VALUES (?, ?, ?, ?, ?)`,
			reportID, rel.Source, rel.Target, rel.Type, string(relProvJson))
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetReports retrieves all parsed reports from the DB
func (db *DB) GetReports() ([]*model.ParsedReport, error) {
	rows, err := db.Conn.Query(`SELECT id, file_name, format, parsed_at FROM reports`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []*model.ParsedReport
	for rows.Next() {
		r := &model.ParsedReport{}
		if err := rows.Scan(&r.ID, &r.FileName, &r.Format, &r.ParsedAt); err != nil {
			return nil, err
		}
		
		// Load components for this report
		comps, err := db.GetComponents(r.ID)
		if err != nil {
			return nil, err
		}
		r.Components = comps
		
		// Load relationships
		rels, err := db.GetRelationships(r.ID)
		if err != nil {
			return nil, err
		}
		r.Relationships = rels

		reports = append(reports, r)
	}
	return reports, nil
}

func (db *DB) GetComponents(reportID int64) ([]model.Component, error) {
	rows, err := db.Conn.Query(`SELECT id, report_id, name, version, purl, supplier, description, download_url, licenses, copyrights, hashes, provenance FROM components WHERE report_id = ?`, reportID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comps []model.Component
	for rows.Next() {
		var c model.Component
		var licJson, copyJson, hashJson, provJson string
		err := rows.Scan(&c.ID, &c.ReportID, &c.Name, &c.Version, &c.PURL, &c.Supplier, &c.Description, &c.DownloadURL, &licJson, &copyJson, &hashJson, &provJson)
		if err != nil {
			return nil, err
		}

		json.Unmarshal([]byte(licJson), &c.Licenses)
		json.Unmarshal([]byte(copyJson), &c.Copyrights)
		json.Unmarshal([]byte(hashJson), &c.Hashes)
		json.Unmarshal([]byte(provJson), &c.Provenance)
		
		comps = append(comps, c)
	}
	return comps, nil
}

func (db *DB) GetRelationships(reportID int64) ([]model.Relationship, error) {
	rows, err := db.Conn.Query(`SELECT source, target, type, provenance FROM relationships WHERE report_id = ?`, reportID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rels []model.Relationship
	for rows.Next() {
		var r model.Relationship
		var provJson string
		if err := rows.Scan(&r.Source, &r.Target, &r.Type, &provJson); err != nil {
			return nil, err
		}
		json.Unmarshal([]byte(provJson), &r.Provenance)
		rels = append(rels, r)
	}
	return rels, nil
}

// LogChange records a manual edit in the changelog table
func (db *DB) LogChange(reportID int64, compName, field, oldVal, newVal, user, reason string) error {
	_, err := db.Conn.Exec(`INSERT INTO changelog (report_id, component_name, field, old_value, new_value, changed_by, reason, timestamp) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		reportID, compName, field, oldVal, newVal, user, reason, time.Now())
	return err
}

// GetChangelog retrieves changes for a report
func (db *DB) GetChangelog(reportID int64) ([]map[string]interface{}, error) {
	rows, err := db.Conn.Query(`SELECT id, component_name, field, old_value, new_value, changed_by, reason, timestamp FROM changelog WHERE report_id = ?`, reportID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []map[string]interface{}
	for rows.Next() {
		var id int64
		var name, field, old, new, user, reason string
		var ts time.Time
		if err := rows.Scan(&id, &name, &field, &old, &new, &user, &reason, &ts); err != nil {
			return nil, err
		}
		logs = append(logs, map[string]interface{}{
			"id":            id,
			"componentName": name,
			"field":         field,
			"oldValue":      old,
			"newValue":      new,
			"changedBy":     user,
			"reason":        reason,
			"timestamp":     ts,
		})
	}
	return logs, nil
}

// ClearAll wipes all data from the database
func (db *DB) ClearAll() error {
	tables := []string{"changelog", "conflicts", "relationships", "components", "reports"}
	for _, t := range tables {
		_, err := db.Conn.Exec(fmt.Sprintf("DELETE FROM %s", t))
		if err != nil {
			return err
		}
	}
	return nil
}

// SaveMergeResult saves a merged report to the DB
func (db *DB) SaveMergeResult(result *model.MergeResult) error {
	tx, err := db.Conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Save the report record (as a special 'merged' report)
	summaryJson, _ := json.Marshal(result.Summary)
	res, err := tx.Exec(`INSERT INTO reports (file_name, format, parsed_at, summary) VALUES (?, ?, ?, ?)`,
		"aggregated-report", "merged", time.Now(), string(summaryJson))
	if err != nil {
		return err
	}
	reportID, _ := res.LastInsertId()
	result.ID = reportID

	// 2. Save components
	for i := range result.Components {
		comp := &result.Components[i]
		comp.ReportID = reportID

		licJson, _ := json.Marshal(comp.Licenses)
		copyJson, _ := json.Marshal(comp.Copyrights)
		hashJson, _ := json.Marshal(comp.Hashes)
		provJson, _ := json.Marshal(comp.Provenance)

		resComp, err := tx.Exec(`INSERT INTO components (report_id, name, version, purl, supplier, description, download_url, licenses, copyrights, hashes, provenance) 
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			reportID, comp.Name, comp.Version, comp.PURL, comp.Supplier, comp.Description, comp.DownloadURL,
			string(licJson), string(copyJson), string(hashJson), string(provJson))
		if err != nil {
			return err
		}
		comp.ID, _ = resComp.LastInsertId()
	}

	// 3. Save relationships
	for _, rel := range result.Relationships {
		relProvJson, _ := json.Marshal(rel.Provenance)
		_, err := tx.Exec(`INSERT INTO relationships (report_id, source, target, type, provenance) VALUES (?, ?, ?, ?, ?)`,
			reportID, rel.Source, rel.Target, rel.Type, string(relProvJson))
		if err != nil {
			return err
		}
	}

	// 4. Save conflicts
	for i := range result.Conflicts {
		conf := &result.Conflicts[i]
		conf.ReportID = reportID

		resConf, err := tx.Exec(`INSERT INTO conflicts (report_id, component_name, component_version, field, value_a, source_a, value_b, source_b, resolved, resolution, resolved_by, resolved_at) 
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			reportID, conf.ComponentName, conf.ComponentVer, conf.Field, conf.ValueA, conf.SourceA, conf.ValueB, conf.SourceB, 
			conf.Resolved, conf.Resolution, conf.ResolvedBy, conf.ResolvedAt)
		if err != nil {
			return err
		}
		conf.ID, _ = resConf.LastInsertId()
	}

	return tx.Commit()
}

// GetLatestMergeResult retrieves the most recent merged report
func (db *DB) GetLatestMergeResult() (*model.MergeResult, error) {
	row := db.Conn.QueryRow(`SELECT id, file_name, format, parsed_at, summary FROM reports WHERE format = 'merged' ORDER BY id DESC LIMIT 1`)
	
	var reportID int64
	var fileName, format string
	var summaryJson sql.NullString
	var parsedAt time.Time
	err := row.Scan(&reportID, &fileName, &format, &parsedAt, &summaryJson)
	if err == sql.ErrNoRows {
		return nil, nil // No merge result yet
	}
	if err != nil {
		return nil, err
	}

	result := &model.MergeResult{
		ID:       reportID,
		MergedAt: parsedAt.Format(time.RFC3339),
	}
	if summaryJson.Valid {
		json.Unmarshal([]byte(summaryJson.String), &result.Summary)
	}

	// Load components
	comps, err := db.GetComponents(reportID)
	if err != nil {
		return nil, err
	}
	result.Components = comps

	// Load relationships
	rels, err := db.GetRelationships(reportID)
	if err != nil {
		return nil, err
	}
	result.Relationships = rels

	// Load conflicts
	conflicts, err := db.GetConflicts(reportID)
	if err != nil {
		return nil, err
	}
	result.Conflicts = conflicts

	return result, nil
}

func (db *DB) GetConflicts(reportID int64) ([]model.Conflict, error) {
	rows, err := db.Conn.Query(`SELECT id, report_id, component_name, component_version, field, value_a, source_a, value_b, source_b, resolved, resolution, resolved_by, resolved_at FROM conflicts WHERE report_id = ?`, reportID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conflicts []model.Conflict
	for rows.Next() {
		var c model.Conflict
		var resolvedAt sql.NullTime
		err := rows.Scan(&c.ID, &c.ReportID, &c.ComponentName, &c.ComponentVer, &c.Field, &c.ValueA, &c.SourceA, &c.ValueB, &c.SourceB, &c.Resolved, &c.Resolution, &c.ResolvedBy, &resolvedAt)
		if err != nil {
			return nil, err
		}
		if resolvedAt.Valid {
			c.ResolvedAt = resolvedAt.Time
		}
		conflicts = append(conflicts, c)
	}
	return conflicts, nil
}

// UpdateConflict updates the resolution status of a conflict
func (db *DB) UpdateConflict(c *model.Conflict) error {
	_, err := db.Conn.Exec(`UPDATE conflicts SET resolved = ?, resolution = ?, resolved_by = ?, resolved_at = ? WHERE id = ?`,
		c.Resolved, c.Resolution, c.ResolvedBy, c.ResolvedAt, c.ID)
	return err
}

// UpdateComponent updates a component's fields in the database
func (db *DB) UpdateComponent(c *model.Component) error {
	licJson, _ := json.Marshal(c.Licenses)
	copyJson, _ := json.Marshal(c.Copyrights)
	hashJson, _ := json.Marshal(c.Hashes)
	provJson, _ := json.Marshal(c.Provenance)

	_, err := db.Conn.Exec(`UPDATE components SET name = ?, version = ?, purl = ?, supplier = ?, description = ?, download_url = ?, licenses = ?, copyrights = ?, hashes = ?, provenance = ? WHERE id = ?`,
		c.Name, c.Version, c.PURL, c.Supplier, c.Description, c.DownloadURL, string(licJson), string(copyJson), string(hashJson), string(provJson), c.ID)
	return err
}

// GetComponentByID retrieves a specific component by its ID
func (db *DB) GetComponentByID(id int64) (*model.Component, error) {
	row := db.Conn.QueryRow(`SELECT id, report_id, name, version, purl, supplier, description, download_url, licenses, copyrights, hashes, provenance FROM components WHERE id = ?`, id)

	var c model.Component
	var licJson, copyJson, hashJson, provJson string
	err := row.Scan(&c.ID, &c.ReportID, &c.Name, &c.Version, &c.PURL, &c.Supplier, &c.Description, &c.DownloadURL, &licJson, &copyJson, &hashJson, &provJson)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(licJson), &c.Licenses)
	json.Unmarshal([]byte(copyJson), &c.Copyrights)
	json.Unmarshal([]byte(hashJson), &c.Hashes)
	json.Unmarshal([]byte(provJson), &c.Provenance)

	return &c, nil
}
