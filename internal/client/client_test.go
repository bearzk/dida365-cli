package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bearzk/dida365-cli/internal/config"
)

func TestNewClient(t *testing.T) {
	cfg := &config.Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		AccessToken:  "test-access-token",
		BaseURL:      "https://api.dida365.com",
	}

	client := NewClient(cfg)

	if client == nil {
		t.Fatal("NewClient returned nil")
	}

	if client.config != cfg {
		t.Errorf("expected config to be %v, got %v", cfg, client.config)
	}

	if client.baseURL != cfg.BaseURL {
		t.Errorf("expected baseURL to be %s, got %s", cfg.BaseURL, client.baseURL)
	}

	if client.httpClient == nil {
		t.Error("expected httpClient to be initialized, got nil")
	}
}

func TestDoRequestAuthHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify Authorization header
		authHeader := r.Header.Get("Authorization")
		expectedAuth := "Bearer test-token"
		if authHeader != expectedAuth {
			t.Errorf("expected Authorization header %s, got %s", expectedAuth, authHeader)
		}

		// Verify Content-Type header
		contentType := r.Header.Get("Content-Type")
		expectedContentType := "application/json"
		if contentType != expectedContentType {
			t.Errorf("expected Content-Type header %s, got %s", expectedContentType, contentType)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	cfg := &config.Config{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		AccessToken:  "test-token",
		BaseURL:      server.URL,
	}

	client := NewClient(cfg)

	var result map[string]string
	err := client.doRequest("GET", "/test", nil, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result["status"] != "ok" {
		t.Errorf("expected status to be ok, got %s", result["status"])
	}
}

func TestDoRequestHTTPErrors(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		responseBody   string
		expectedError  string
	}{
		{
			name:          "401 Unauthorized",
			statusCode:    http.StatusUnauthorized,
			responseBody:  `{"error": "invalid token"}`,
			expectedError: "access token expired or invalid",
		},
		{
			name:          "403 Forbidden",
			statusCode:    http.StatusForbidden,
			responseBody:  `{"error": "forbidden"}`,
			expectedError: "insufficient permissions for this operation",
		},
		{
			name:          "404 Not Found",
			statusCode:    http.StatusNotFound,
			responseBody:  `{"error": "not found"}`,
			expectedError: "resource not found",
		},
		{
			name:          "500 Internal Server Error",
			statusCode:    http.StatusInternalServerError,
			responseBody:  `{"error": "server error"}`,
			expectedError: "Dida365 server error, try again later",
		},
		{
			name:          "400 Bad Request with error message",
			statusCode:    http.StatusBadRequest,
			responseBody:  `{"errorMessage": "invalid input data"}`,
			expectedError: "invalid input data",
		},
		{
			name:          "400 Bad Request without error message",
			statusCode:    http.StatusBadRequest,
			responseBody:  `{"error": "bad request"}`,
			expectedError: "bad request (status 400)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			cfg := &config.Config{
				ClientID:     "test-client",
				ClientSecret: "test-secret",
				AccessToken:  "test-token",
				BaseURL:      server.URL,
			}

			client := NewClient(cfg)

			var result map[string]string
			err := client.doRequest("GET", "/test", nil, &result)
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if err.Error() != tt.expectedError {
				t.Errorf("expected error %q, got %q", tt.expectedError, err.Error())
			}
		})
	}
}

func TestDoRequestJSONMarshaling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request body was marshaled correctly
		var reqBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		if reqBody["name"] != "Test Task" {
			t.Errorf("expected name to be 'Test Task', got %v", reqBody["name"])
		}

		if reqBody["priority"] != float64(1) {
			t.Errorf("expected priority to be 1, got %v", reqBody["priority"])
		}

		// Return response to be unmarshaled
		response := map[string]interface{}{
			"id":       "123",
			"name":     "Test Task",
			"priority": 1,
			"status":   "active",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	cfg := &config.Config{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		AccessToken:  "test-token",
		BaseURL:      server.URL,
	}

	client := NewClient(cfg)

	requestBody := map[string]interface{}{
		"name":     "Test Task",
		"priority": 1,
	}

	var result map[string]interface{}
	err := client.doRequest("POST", "/tasks", requestBody, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result["id"] != "123" {
		t.Errorf("expected id to be '123', got %v", result["id"])
	}

	if result["status"] != "active" {
		t.Errorf("expected status to be 'active', got %v", result["status"])
	}
}
