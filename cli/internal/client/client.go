package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/vknow360/otaship/cli/internal/utils"
)

type Client struct {
	BaseURL string
}

type ValidateKeyResponse struct {
	ProjectID string `json:"project_id"`
	Slug      string `json:"slug"`
	Name      string `json:"name"`
}

type ProjectDetails struct {
	ID   string `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type CreateUpdateRequest struct {
	ID                string `json:"id"`
	ProjectID         string `json:"project_id"`
	RuntimeVersion    string `json:"runtime_version"`
	Channel           string `json:"channel"`
	Platform          string `json:"platform"`
	RolloutPercentage int    `json:"rollout_percentage"`
}

type CreateUpdateResponse struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
	Platform  string `json:"platform"`
}

func (c *Client) ValidateAPIKey(apiKey string) (*ValidateKeyResponse, error) {
	req, _ := http.NewRequest(
		"GET",
		c.BaseURL+"/api/validate-key",
		nil)
	req.Header.Set("X-API-Key", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, utils.HandleHTTPError(resp)
	}

	var result ValidateKeyResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) GetProjectByID(apiKey string) (*ProjectDetails, error) {
	url := fmt.Sprintf(
		"%s/api/project/me", c.BaseURL,
	)

	httpReq, _ := http.NewRequest("GET", url, nil)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-API-Key", apiKey)
	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, utils.HandleHTTPError(resp)
	}

	var result ProjectDetails
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) CreateUpdate(apiKey string, req *CreateUpdateRequest) (*CreateUpdateResponse, error) {
	url := fmt.Sprintf("%s/api/project/updates", c.BaseURL)

	body, _ := json.Marshal(req)
	httpReq, _ := http.NewRequest("POST", url, bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-API-Key", apiKey)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		return nil, utils.HandleHTTPError(resp)
	}

	var result CreateUpdateResponse
	json.NewDecoder(resp.Body).Decode(&result)
	return &result, nil
}

func (c *Client) UploadBundle(projectID, updateID, platform, apiKey, zipPath string) error {
	url := fmt.Sprintf("%s/api/project/%s/updates/%s/upload",
		c.BaseURL, projectID, updateID)

	file, err := os.Open(zipPath)
	if err != nil {
		return err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	writer.WriteField("platform", platform)

	part, _ := writer.CreateFormFile("bundle", filepath.Base(zipPath))
	io.Copy(part, file)
	writer.Close()

	req, _ := http.NewRequest("POST", url, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-API-Key", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("upload failed with status %d", resp.StatusCode)
	}
	return nil
}

type UpdateSummary struct {
	ID                string `json:"id"`
	ProjectID         string `json:"project_id"`
	RuntimeVersion    string `json:"runtime_version"`
	Channel           string `json:"channel"`
	Platform          string `json:"platform"`
	RolloutPercentage int    `json:"rollout_percentage"`
	IsActive          bool   `json:"is_active"`
	IsRollback        bool   `json:"is_rollback"`
	Message           string `json:"message"`
	CreatedAt         string `json:"created_at"`
}

func (c *Client) ListUpdates(apiKey string) ([]UpdateSummary, error) {
	req, _ := http.NewRequest("GET", c.BaseURL+"/api/project/updates", nil)
	req.Header.Set("X-API-Key", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, utils.HandleHTTPError(resp)
	}

	var updates []UpdateSummary
	json.NewDecoder(resp.Body).Decode(&updates)
	return updates, nil
}

func (c *Client) DeleteUpdate(apiKey, updateID string) error {
	url := fmt.Sprintf("%s/api/project/updates/%s", c.BaseURL, updateID)
	req, _ := http.NewRequest("DELETE", url, nil)
	req.Header.Set("X-API-Key", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 {
		return utils.HandleHTTPError(resp)
	}
	return nil
}

func (c *Client) RollbackUpdate(apiKey, updateID string) (*CreateUpdateResponse, error) {
	url := fmt.Sprintf("%s/api/project/updates/%s/rollback", c.BaseURL, updateID)
	req, _ := http.NewRequest("POST", url, nil)
	req.Header.Set("X-API-Key", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		return nil, utils.HandleHTTPError(resp)
	}

	var result CreateUpdateResponse
	json.NewDecoder(resp.Body).Decode(&result)
	return &result, nil
}
