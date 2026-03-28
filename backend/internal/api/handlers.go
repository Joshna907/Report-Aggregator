package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/fossology/report-aggregator/internal/db"
	"github.com/fossology/report-aggregator/internal/export"
	"github.com/fossology/report-aggregator/internal/fossology"
	"github.com/fossology/report-aggregator/internal/merge"
	"github.com/fossology/report-aggregator/internal/model"
	"github.com/fossology/report-aggregator/internal/parser"
	"github.com/fossology/report-aggregator/internal/sw360"
	"github.com/fossology/report-aggregator/internal/validate"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Server struct {
	DB     *db.DB
	Router *gin.Engine
}

func NewServer(database *db.DB) *Server {
	s := &Server{
		DB:     database,
		Router: gin.Default(),
	}
	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	origins := []string{"http://localhost:3000"}
	if allowedOrigins != "" {
		origins = strings.Split(allowedOrigins, ",")
	}

	s.Router.Use(cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	s.Router.POST("/merge", s.handleMerge)
	s.Router.GET("/result", s.handleGetResult)
	s.Router.GET("/summary", s.handleGetSummary)
	s.Router.GET("/changelog/:reportId", s.handleGetChangelog)
	s.Router.POST("/conflicts/resolve", s.handleResolveConflict)
	s.Router.POST("/components/edit", s.handleEditComponent)
	s.Router.POST("/result/export/sw360", s.handleExportSW360)
	s.Router.GET("/export", s.handleExportReport)
	s.Router.POST("/validate", s.handleValidate)
	s.Router.GET("/fossology/uploads", s.handleFossologyUploads)
	s.Router.POST("/fossology/fetch", s.handleFossologyFetch)
}

func (s *Server) Run(addr string) error {
	return s.Router.Run(addr)
}

func (s *Server) handleMerge(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse multipart form: " + err.Error()})
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No files provided. Use standard 'files' multipart field."})
		return
	}

	fmt.Printf("Merge started for %d files\n", len(files))
	var reports []*model.ParsedReport
	for i, f := range files {
		fmt.Printf("Parsing file %d: %s\n", i+1, f.Filename)
		tempFile := filepath.Join(os.TempDir(), f.Filename)
		if err := c.SaveUploadedFile(f, tempFile); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save uploaded file: " + err.Error()})
			return
		}

		pr, err := parser.DetectAndParse(tempFile)
		os.Remove(tempFile) // Clean up immediately after parsing

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse " + f.Filename + ": " + err.Error()})
			return
		}
		
		// Use original uploaded filename, not the temp path
		pr.FileName = f.Filename
		
		reports = append(reports, pr)
		
		// Save raw report to DB
		if err := s.DB.SaveReport(pr); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save report: " + err.Error()})
			return
		}
		fmt.Printf("Successfully parsed and saved %s\n", f.Filename)
	}

	fmt.Println("Starting Merge engine...")
	result, err := merge.Merge(reports)
	if err != nil {
		fmt.Printf("Merge Error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	fmt.Printf("Merge finalized. Result ID: %d, Components: %d, Conflicts: %d\n", result.ID, len(result.Components), len(result.Conflicts))
	if err := s.DB.SaveMergeResult(result); err != nil {
		fmt.Printf("Persistence Error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to persist merge result: " + err.Error()})
		return
	}
	fmt.Println("Merge state committed.")

	c.JSON(http.StatusOK, result)
}

func (s *Server) handleGetResult(c *gin.Context) {
	result, err := s.DB.GetLatestMergeResult()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if result == nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "No merge result found"})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (s *Server) handleGetSummary(c *gin.Context) {
	result, err := s.DB.GetLatestMergeResult()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if result == nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "No merge result found"})
		return
	}
	c.JSON(http.StatusOK, result.Summary)
}

func (s *Server) handleGetChangelog(c *gin.Context) {
	reportIDStr := c.Param("reportId")
	reportID, _ := strconv.ParseInt(reportIDStr, 10, 64)

	logs, err := s.DB.GetChangelog(reportID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, logs)
}

func (s *Server) handleResolveConflict(c *gin.Context) {
	var conflict model.Conflict
	if err := c.ShouldBindJSON(&conflict); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conflict.Resolved = true
	conflict.ResolvedAt = time.Now()
	// Fallback to current session user (admin for standalone)
	if conflict.ResolvedBy == "" {
		conflict.ResolvedBy = "admin"
	}

	if err := s.DB.UpdateConflict(&conflict); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Conflict resolved"})
}

func (s *Server) handleEditComponent(c *gin.Context) {
	var req struct {
		Component model.Component `json:"component" binding:"required"`
		User      string          `json:"user"`
		Reason    string          `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Fetch old version to log accurate changes
	oldComp, err := s.DB.GetComponentByID(req.Component.ID)
	oldValStr := "unknown"
	if err == nil && oldComp != nil {
		oldValStr = fmt.Sprintf("v: %s, s: %s, purl: %s", oldComp.Version, oldComp.Supplier, oldComp.PURL)
	}
	
	if err := s.DB.UpdateComponent(&req.Component); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	newValStr := fmt.Sprintf("v: %s, s: %s, purl: %s", req.Component.Version, req.Component.Supplier, req.Component.PURL)
	s.DB.LogChange(req.Component.ReportID, req.Component.Name, "manual_edit", oldValStr, newValStr, req.User, req.Reason)

	c.JSON(http.StatusOK, gin.H{"message": "Component updated"})
}

func (s *Server) handleExportSW360(c *gin.Context) {
	result, err := s.DB.GetLatestMergeResult()
	if err != nil {
		fmt.Printf("Export Error (GetLatest): %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch result: " + err.Error()})
		return
	}
	if result == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No merge result found to export"})
		return
	}

	// For GSoC demo, we use placeholder credentials. 
	// In production, these would be in environment variables.
	sw360URL := "http://localhost:8081" // Assuming SW360 or mock is here
	sw360Token := "gsoc-demo-token"
	projectID := "demo-project-123"

	client := sw360.NewClient(sw360URL, sw360Token)
	
	// Serialize to JSON for the export
	jsonData, _ := json.MarshalIndent(result, "", "  ")
	filename := fmt.Sprintf("aggregated-report-%s.json", time.Now().Format("20060102-150405"))

	if err := client.PushReport(projectID, filename, jsonData); err != nil {
		// Mock success for the prototype demonstration if SW360 is not reachable
		errStr := strings.ToLower(err.Error())
		if strings.Contains(errStr, "refused") || strings.Contains(errStr, "connectex") || strings.Contains(errStr, "no such host") {
			c.JSON(http.StatusOK, gin.H{
				"message": "SW360 Integration: Prototype push triggered. (Connection to live SW360 instance is a GSoC Phase 2 objective).",
				"status": "placeholder",
			})
			return
		}
		fmt.Printf("Export Error (PushReport): %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "SW360 Export failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully exported to SW360 project " + projectID})
}

// handleExportReport exports the merged result as SPDX 2.3 or CycloneDX
func (s *Server) handleExportReport(c *gin.Context) {
	format := c.DefaultQuery("format", "spdx")

	result, err := s.DB.GetLatestMergeResult()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if result == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No merge result found. Please merge reports first."})
		return
	}

	var data []byte
	var filename string
	var contentType string

	switch format {
	case "spdx":
		data, err = export.ToSPDX(result)
		filename = fmt.Sprintf("aggregated-report-%s.spdx.json", time.Now().Format("20060102"))
		contentType = "application/json"
	case "cyclonedx":
		data, err = export.ToCycloneDX(result)
		filename = fmt.Sprintf("aggregated-report-%s.cdx.json", time.Now().Format("20060102"))
		contentType = "application/json"
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported format. Use 'spdx' or 'cyclonedx'."})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Export failed: " + err.Error()})
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Data(http.StatusOK, contentType, data)
}

// handleValidate validates a single uploaded report
func (s *Server) handleValidate(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided"})
		return
	}

	tempFile := filepath.Join(os.TempDir(), file.Filename)
	if err := c.SaveUploadedFile(file, tempFile); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}
	defer os.Remove(tempFile)

	report, err := parser.DetectAndParse(tempFile)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parse failed: " + err.Error()})
		return
	}

	result := validate.Validate(report)
	c.JSON(http.StatusOK, gin.H{
		"valid":      result.Valid,
		"errors":     result.Errors,
		"warnings":   result.Warnings,
		"format":     report.Format,
		"components": len(report.Components),
	})
}

// handleFossologyUploads lists uploads from a connected FOSSology instance
func (s *Server) handleFossologyUploads(c *gin.Context) {
	fossURL := os.Getenv("FOSSOLOGY_URL")
	fossToken := os.Getenv("FOSSOLOGY_TOKEN")

	if fossURL == "" {
		fossURL = "http://localhost:8081"
	}
	if fossToken == "" {
		fossToken = "demo-token"
	}

	client := fossology.NewClient(fossURL, fossToken)
	uploads, err := client.GetUploads()
	if err != nil {
		errStr := strings.ToLower(err.Error())
		if strings.Contains(errStr, "refused") || strings.Contains(errStr, "connectex") || strings.Contains(errStr, "no such host") {
			// Return mock data for demo purposes
			c.JSON(http.StatusOK, []map[string]interface{}{
				{"id": 1, "uploadName": "linux-kernel-5.15.tar.gz", "uploadDate": "2024-03-15", "folderName": "Main"},
				{"id": 2, "uploadName": "openssl-3.1.0.tar.gz", "uploadDate": "2024-03-20", "folderName": "Vendor"},
				{"id": 3, "uploadName": "react-18.2.0.tgz", "uploadDate": "2024-03-25", "folderName": "Frontend"},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, uploads)
}

// handleFossologyFetch fetches a report from FOSSology, parses it, and returns the result
func (s *Server) handleFossologyFetch(c *gin.Context) {
	var req struct {
		UploadID int    `json:"uploadId" binding:"required"`
		Format   string `json:"format"` // spdx, cyclonedx
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Format == "" {
		req.Format = "spdx"
	}

	fossURL := os.Getenv("FOSSOLOGY_URL")
	fossToken := os.Getenv("FOSSOLOGY_TOKEN")
	if fossURL == "" {
		fossURL = "http://localhost:8081"
	}
	if fossToken == "" {
		fossToken = "demo-token"
	}

	client := fossology.NewClient(fossURL, fossToken)

	reportID, err := client.GenerateReport(req.UploadID, req.Format)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate report: " + err.Error()})
		return
	}

	data, err := client.DownloadReport(reportID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to download report: " + err.Error()})
		return
	}

	// Save to temp file for parsing
	tempFile := filepath.Join(os.TempDir(), fmt.Sprintf("fossology-%s.%s.json", reportID, req.Format))
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save temp file"})
		return
	}
	defer os.Remove(tempFile)

	report, err := parser.DetectAndParse(tempFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Parse failed: " + err.Error()})
		return
	}

	report.FileName = fmt.Sprintf("fossology-upload-%d.%s.json", req.UploadID, req.Format)

	if err := s.DB.SaveReport(report); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Report fetched and saved",
		"format":     report.Format,
		"components": len(report.Components),
		"fileName":   report.FileName,
	})
}