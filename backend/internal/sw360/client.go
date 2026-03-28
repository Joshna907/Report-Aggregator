package sw360

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

// Client for interacting with SW360 REST API
type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

// NewClient creates a new SW360 API client
func NewClient(baseURL, token string) *Client {
	return &Client{
		BaseURL: baseURL,
		Token:   token,
		HTTPClient: &http.Client{
			Timeout: time.Second * 30,
		},
	}
}

// Project represents a SW360 project
type Project struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
}

// Attachment represents a file attached to a project/release in SW360
type Attachment struct {
	Filename       string `json:"filename"`
	AttachmentType string `json:"attachmentType"` // e.g., "DOCUMENT"
	Links          struct {
		Download struct {
			Href string `json:"href"`
		} `json:"self"` // SW360 sometimes uses 'self' or 'download'
	} `json:"_links"`
}

// GetProjects retrieves a list of projects from SW360
func (c *Client) GetProjects() ([]Project, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/projects", c.BaseURL), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Token "+c.Token)
	req.Header.Set("Accept", "application/hal+json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch projects: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("SW360 API error: status %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)

	// SW360 uses HAL JSON format
	var result struct {
		Embedded struct {
			Projects []Project `json:"sw360:projects"`
		} `json:"_embedded"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse SW360 response: %w", err)
	}

	return result.Embedded.Projects, nil
}

// GetProjectAttachments gets all attachments (including reports) for a project
func (c *Client) GetProjectAttachments(projectID string) ([]Attachment, error) {
	url := fmt.Sprintf("%s/api/projects/%s/attachments", c.BaseURL, projectID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Token "+c.Token)
	req.Header.Set("Accept", "application/hal+json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get project attachments, status %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Embedded struct {
			Attachments []Attachment `json:"sw360:attachments"`
		} `json:"_embedded"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	// For MVP, ensure we have the Href available.
	// SW360 HAL structure for attachments can be complex.
	return result.Embedded.Attachments, nil
}

// DownloadAttachment downloads a specific report/attachment
func (c *Client) DownloadAttachment(downloadURL string) ([]byte, error) {
	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Token "+c.Token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download attachment, status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// PushReport uploads the aggregated report back to a SW360 project
func (c *Client) PushReport(projectID, filename string, data []byte) error {
	url := fmt.Sprintf("%s/api/projects/%s/attachments", c.BaseURL, projectID)

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return err
	}
	part.Write(data)

	// SW360 standard attachment metadata
	writer.WriteField("attachmentInfo", `{"attachmentType":"DOCUMENT"}`)
	writer.Close()

	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Token "+c.Token)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to push report, status %d", resp.StatusCode)
	}

	return nil
}
