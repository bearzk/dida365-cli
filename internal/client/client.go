package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/bearzk/dida365-cli/internal/config"
)

// Error variables for typed error handling
var (
	ErrUnauthorized = fmt.Errorf("unauthorized: access token expired or invalid")
	ErrForbidden    = fmt.Errorf("forbidden: insufficient permissions")
	ErrNotFound     = fmt.Errorf("not found: resource does not exist")
)

// Client represents an HTTP client for communicating with the Dida365 API
type Client struct {
	httpClient *http.Client
	config     *config.Config
	baseURL    string
}

// NewClient creates a new Dida365 API client
func NewClient(cfg *config.Config) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		config:  cfg,
		baseURL: cfg.BaseURL,
	}
}

// doRequest performs an HTTP request with authentication and JSON handling
func (c *Client) doRequest(method, path string, body, result interface{}) error {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	url := c.baseURL + path
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+c.config.AccessToken)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Handle HTTP errors
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return c.handleHTTPError(resp.StatusCode, respBody)
	}

	// Unmarshal response if result is provided (skip if body is empty)
	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// handleHTTPError converts HTTP error responses to Go errors
func (c *Client) handleHTTPError(statusCode int, body []byte) error {
	switch statusCode {
	case http.StatusUnauthorized:
		return fmt.Errorf("%w", ErrUnauthorized)
	case http.StatusForbidden:
		return fmt.Errorf("%w", ErrForbidden)
	case http.StatusNotFound:
		return fmt.Errorf("%w", ErrNotFound)
	case http.StatusBadRequest:
		// Try to parse error message from response body
		var apiError struct {
			ErrorMessage string `json:"errorMessage"`
		}
		if err := json.Unmarshal(body, &apiError); err == nil && apiError.ErrorMessage != "" {
			return fmt.Errorf("%s", apiError.ErrorMessage)
		}
		// If no error message found, return generic error
		return fmt.Errorf("bad request (status %d)", statusCode)
	case http.StatusInternalServerError:
		return fmt.Errorf("Dida365 server error, try again later")
	default:
		return fmt.Errorf("HTTP error: status %d", statusCode)
	}
}
