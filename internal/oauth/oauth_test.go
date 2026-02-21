package oauth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestExchangeCodeForToken(t *testing.T) {
	tests := []struct {
		name           string
		responseStatus int
		responseBody   string
		wantErr        bool
		errContains    string
	}{
		{
			name:           "success",
			responseStatus: http.StatusOK,
			responseBody: `{
				"access_token": "test_access_token",
				"refresh_token": "test_refresh_token",
				"expires_in": 3600,
				"token_type": "Bearer"
			}`,
			wantErr: false,
		},
		{
			name:           "error response",
			responseStatus: http.StatusBadRequest,
			responseBody: `{
				"error": "invalid_grant",
				"error_description": "Invalid authorization code"
			}`,
			wantErr:     true,
			errContains: "invalid_grant",
		},
		{
			name:           "invalid json response",
			responseStatus: http.StatusOK,
			responseBody:   `not json`,
			wantErr:        true,
			errContains:    "failed to decode token response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify POST method
				if r.Method != http.MethodPost {
					t.Errorf("Expected POST method, got %s", r.Method)
				}

				// Verify Content-Type header
				contentType := r.Header.Get("Content-Type")
				if contentType != "application/x-www-form-urlencoded" {
					t.Errorf("Expected Content-Type application/x-www-form-urlencoded, got %s", contentType)
				}

				// Parse form data
				if err := r.ParseForm(); err != nil {
					t.Fatalf("Failed to parse form: %v", err)
				}

				// Verify required parameters
				if r.FormValue("grant_type") != "authorization_code" {
					t.Errorf("Expected grant_type=authorization_code, got %s", r.FormValue("grant_type"))
				}
				if r.FormValue("code") != "test_code" {
					t.Errorf("Expected code=test_code, got %s", r.FormValue("code"))
				}
				if r.FormValue("client_id") != "test_client_id" {
					t.Errorf("Expected client_id=test_client_id, got %s", r.FormValue("client_id"))
				}
				if r.FormValue("client_secret") != "test_client_secret" {
					t.Errorf("Expected client_secret=test_client_secret, got %s", r.FormValue("client_secret"))
				}
				if r.FormValue("redirect_uri") != "http://localhost:8080/callback" {
					t.Errorf("Expected redirect_uri=http://localhost:8080/callback, got %s", r.FormValue("redirect_uri"))
				}

				// Send response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.responseStatus)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			// Test the function
			tokenResp, err := exchangeCodeForToken(
				server.URL,
				"test_code",
				"test_client_id",
				"test_client_secret",
				"http://localhost:8080/callback",
			)

			// Check error
			if tt.wantErr {
				if err == nil {
					t.Fatalf("Expected error containing %q, got nil", tt.errContains)
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Expected error containing %q, got %q", tt.errContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Check response
			if tokenResp.AccessToken != "test_access_token" {
				t.Errorf("Expected AccessToken=test_access_token, got %s", tokenResp.AccessToken)
			}
			if tokenResp.RefreshToken != "test_refresh_token" {
				t.Errorf("Expected RefreshToken=test_refresh_token, got %s", tokenResp.RefreshToken)
			}
			if tokenResp.ExpiresIn != 3600 {
				t.Errorf("Expected ExpiresIn=3600, got %d", tokenResp.ExpiresIn)
			}
			if tokenResp.TokenType != "Bearer" {
				t.Errorf("Expected TokenType=Bearer, got %s", tokenResp.TokenType)
			}
		})
	}
}

func TestRefreshAccessToken(t *testing.T) {
	tests := []struct {
		name           string
		responseStatus int
		responseBody   string
		wantErr        bool
		errContains    string
	}{
		{
			name:           "success",
			responseStatus: http.StatusOK,
			responseBody: `{
				"access_token": "new_access_token",
				"refresh_token": "new_refresh_token",
				"expires_in": 7200,
				"token_type": "Bearer"
			}`,
			wantErr: false,
		},
		{
			name:           "error response",
			responseStatus: http.StatusBadRequest,
			responseBody: `{
				"error": "invalid_grant",
				"error_description": "Invalid refresh token"
			}`,
			wantErr:     true,
			errContains: "invalid_grant",
		},
		{
			name:           "network error",
			responseStatus: http.StatusInternalServerError,
			responseBody:   `Internal Server Error`,
			wantErr:        true,
			errContains:    "token request failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify POST method
				if r.Method != http.MethodPost {
					t.Errorf("Expected POST method, got %s", r.Method)
				}

				// Verify Content-Type header
				contentType := r.Header.Get("Content-Type")
				if contentType != "application/x-www-form-urlencoded" {
					t.Errorf("Expected Content-Type application/x-www-form-urlencoded, got %s", contentType)
				}

				// Parse form data
				if err := r.ParseForm(); err != nil {
					t.Fatalf("Failed to parse form: %v", err)
				}

				// Verify required parameters
				if r.FormValue("grant_type") != "refresh_token" {
					t.Errorf("Expected grant_type=refresh_token, got %s", r.FormValue("grant_type"))
				}
				if r.FormValue("refresh_token") != "old_refresh_token" {
					t.Errorf("Expected refresh_token=old_refresh_token, got %s", r.FormValue("refresh_token"))
				}
				if r.FormValue("client_id") != "test_client_id" {
					t.Errorf("Expected client_id=test_client_id, got %s", r.FormValue("client_id"))
				}
				if r.FormValue("client_secret") != "test_client_secret" {
					t.Errorf("Expected client_secret=test_client_secret, got %s", r.FormValue("client_secret"))
				}

				// Send response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.responseStatus)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			// Test the function
			tokenResp, err := refreshAccessToken(
				server.URL,
				"old_refresh_token",
				"test_client_id",
				"test_client_secret",
			)

			// Check error
			if tt.wantErr {
				if err == nil {
					t.Fatalf("Expected error containing %q, got nil", tt.errContains)
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Expected error containing %q, got %q", tt.errContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Check response
			if tokenResp.AccessToken != "new_access_token" {
				t.Errorf("Expected AccessToken=new_access_token, got %s", tokenResp.AccessToken)
			}
			if tokenResp.RefreshToken != "new_refresh_token" {
				t.Errorf("Expected RefreshToken=new_refresh_token, got %s", tokenResp.RefreshToken)
			}
			if tokenResp.ExpiresIn != 7200 {
				t.Errorf("Expected ExpiresIn=7200, got %d", tokenResp.ExpiresIn)
			}
			if tokenResp.TokenType != "Bearer" {
				t.Errorf("Expected TokenType=Bearer, got %s", tokenResp.TokenType)
			}
		})
	}
}

func TestRequestToken(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token":  "token123",
				"refresh_token": "refresh123",
				"expires_in":    3600,
				"token_type":    "Bearer",
			})
		}))
		defer server.Close()

		data := url.Values{}
		data.Set("grant_type", "authorization_code")

		resp, err := requestToken(server.URL, data)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if resp.AccessToken != "token123" {
			t.Errorf("Expected AccessToken=token123, got %s", resp.AccessToken)
		}
	})

	t.Run("error response with json", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":             "invalid_request",
				"error_description": "Missing required parameter",
			})
		}))
		defer server.Close()

		data := url.Values{}
		_, err := requestToken(server.URL, data)
		if err == nil {
			t.Fatal("Expected error, got nil")
		}

		if !strings.Contains(err.Error(), "invalid_request") {
			t.Errorf("Expected error to contain 'invalid_request', got: %v", err)
		}
	})

	t.Run("invalid url", func(t *testing.T) {
		data := url.Values{}
		_, err := requestToken("http://invalid-url-that-does-not-exist-12345.local", data)
		if err == nil {
			t.Fatal("Expected error for invalid URL, got nil")
		}
	})
}

