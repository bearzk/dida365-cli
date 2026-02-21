# OAuth2 Authentication Flow Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement this plan task-by-task.

**Goal:** Implement standard OAuth2 authorization code flow with browser-based authentication, automatic token exchange, and refresh token management.

**Architecture:** Create new `internal/oauth/` package with flow orchestrator, callback server, and browser opener. Update `cmd/auth.go` to replace `auth configure` with `auth login` command. Extend config to store refresh tokens and expiry.

**Tech Stack:** Go stdlib (net/http, os/exec, crypto/rand), Cobra CLI framework

---

## Task 1: Update Config with Token Fields

**Files:**
- Modify: `internal/config/config.go`
- Modify: `internal/config/config_test.go`

**Step 1: Write tests for new config fields**

Add to `internal/config/config_test.go`:

```go
func TestConfigWithTokenFields(t *testing.T) {
	t.Run("marshal config with refresh token and expiry", func(t *testing.T) {
		expiry := time.Date(2026, 2, 21, 16, 30, 0, 0, time.UTC)
		cfg := &Config{
			ClientID:     "test_client",
			ClientSecret: "test_secret",
			AccessToken:  "test_access",
			BaseURL:      "https://dida365.com",
			RefreshToken: "test_refresh",
			TokenExpiry:  expiry,
		}

		data, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			t.Fatalf("failed to marshal: %v", err)
		}

		// Verify JSON contains new fields
		jsonStr := string(data)
		if !strings.Contains(jsonStr, "refresh_token") {
			t.Error("JSON missing refresh_token field")
		}
		if !strings.Contains(jsonStr, "token_expiry") {
			t.Error("JSON missing token_expiry field")
		}
	})

	t.Run("unmarshal config with token fields", func(t *testing.T) {
		jsonData := `{
  "client_id": "test",
  "client_secret": "secret",
  "access_token": "access",
  "base_url": "https://dida365.com",
  "refresh_token": "refresh",
  "token_expiry": "2026-02-21T16:30:00Z"
}`

		var cfg Config
		err := json.Unmarshal([]byte(jsonData), &cfg)
		if err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}

		if cfg.RefreshToken != "refresh" {
			t.Errorf("RefreshToken: got %s, want refresh", cfg.RefreshToken)
		}

		expectedExpiry := time.Date(2026, 2, 21, 16, 30, 0, 0, time.UTC)
		if !cfg.TokenExpiry.Equal(expectedExpiry) {
			t.Errorf("TokenExpiry: got %v, want %v", cfg.TokenExpiry, expectedExpiry)
		}
	})

	t.Run("backward compatibility - old config without token fields", func(t *testing.T) {
		jsonData := `{
  "client_id": "test",
  "client_secret": "secret",
  "access_token": "access",
  "base_url": "https://dida365.com"
}`

		var cfg Config
		err := json.Unmarshal([]byte(jsonData), &cfg)
		if err != nil {
			t.Fatalf("failed to unmarshal old config: %v", err)
		}

		if cfg.RefreshToken != "" {
			t.Error("RefreshToken should be empty for old config")
		}

		if !cfg.TokenExpiry.IsZero() {
			t.Error("TokenExpiry should be zero for old config")
		}
	})
}

func TestConfigIsExpired(t *testing.T) {
	tests := []struct {
		name    string
		expiry  time.Time
		want    bool
	}{
		{"no expiry set", time.Time{}, false},
		{"expired token", time.Now().Add(-1 * time.Hour), true},
		{"valid token", time.Now().Add(1 * time.Hour), false},
		{"just expired", time.Now().Add(-1 * time.Second), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				TokenExpiry: tt.expiry,
			}
			if got := cfg.IsExpired(); got != tt.want {
				t.Errorf("IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigCanRefresh(t *testing.T) {
	tests := []struct {
		name         string
		refreshToken string
		want         bool
	}{
		{"has refresh token", "refresh_token_123", true},
		{"no refresh token", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				RefreshToken: tt.refreshToken,
			}
			if got := cfg.CanRefresh(); got != tt.want {
				t.Errorf("CanRefresh() = %v, want %v", got, tt.want)
			}
		})
	}
}
```

**Step 2: Run tests to verify they fail**

```bash
go test ./internal/config -v -run "TestConfigWithTokenFields|TestConfigIsExpired|TestConfigCanRefresh"
```

Expected: FAIL with "undefined: Config.RefreshToken", "undefined: Config.TokenExpiry", "undefined: Config.IsExpired", "undefined: Config.CanRefresh"

**Step 3: Add new fields and methods to Config**

Update `internal/config/config.go`:

```go
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Config represents the application configuration
type Config struct {
	ClientID     string    `json:"client_id"`
	ClientSecret string    `json:"client_secret"`
	AccessToken  string    `json:"access_token"`
	BaseURL      string    `json:"base_url"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	TokenExpiry  time.Time `json:"token_expiry,omitempty"`
}

// ... existing Load, Save, Validate, DefaultConfigPath functions ...

// IsExpired checks if the access token has expired
// Returns false if no expiry is set (backward compatibility)
func (c *Config) IsExpired() bool {
	if c.TokenExpiry.IsZero() {
		return false
	}
	return time.Now().After(c.TokenExpiry)
}

// CanRefresh checks if the config has a refresh token available
func (c *Config) CanRefresh() bool {
	return c.RefreshToken != ""
}
```

**Step 4: Run tests to verify they pass**

```bash
go test ./internal/config -v
```

Expected: PASS for all tests including new ones

**Step 5: Commit**

```bash
git add internal/config/config.go internal/config/config_test.go
git commit -m "feat(config): add refresh token and expiry fields with backward compatibility"
```

---

## Task 2: Implement Browser Opener

**Files:**
- Create: `internal/oauth/browser.go`
- Create: `internal/oauth/browser_test.go`

**Step 1: Write tests for browser opener**

Create `internal/oauth/browser_test.go`:

```go
package oauth

import (
	"runtime"
	"testing"
)

func TestGetBrowserCommand(t *testing.T) {
	tests := []struct {
		goos    string
		wantCmd string
	}{
		{"darwin", "open"},
		{"linux", "xdg-open"},
		{"windows", "cmd"},
	}

	for _, tt := range tests {
		t.Run(tt.goos, func(t *testing.T) {
			cmd, _ := getBrowserCommand(tt.goos, "https://example.com")
			if cmd.Path != tt.wantCmd && cmd.Args[0] != tt.wantCmd {
				t.Errorf("getBrowserCommand(%s) cmd = %v, want command starting with %s", tt.goos, cmd, tt.wantCmd)
			}
		})
	}
}

func TestBuildAuthURL(t *testing.T) {
	tests := []struct {
		name     string
		service  string
		clientID string
		redirect string
		state    string
		wantHost string
	}{
		{
			name:     "dida365 service",
			service:  "dida365",
			clientID: "test_client",
			redirect: "http://localhost:8080/callback",
			state:    "random_state",
			wantHost: "dida365.com",
		},
		{
			name:     "ticktick service",
			service:  "ticktick",
			clientID: "test_client",
			redirect: "http://localhost:8080/callback",
			state:    "random_state",
			wantHost: "ticktick.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := buildAuthURL(tt.service, tt.clientID, tt.redirect, tt.state)
			if !strings.Contains(url, tt.wantHost) {
				t.Errorf("buildAuthURL() host = %s, want %s in URL", url, tt.wantHost)
			}
			if !strings.Contains(url, "client_id="+tt.clientID) {
				t.Error("URL missing client_id parameter")
			}
			if !strings.Contains(url, "redirect_uri=") {
				t.Error("URL missing redirect_uri parameter")
			}
			if !strings.Contains(url, "state="+tt.state) {
				t.Error("URL missing state parameter")
			}
			if !strings.Contains(url, "response_type=code") {
				t.Error("URL missing response_type=code")
			}
		})
	}
}
```

**Step 2: Run tests to verify they fail**

```bash
go test ./internal/oauth -v
```

Expected: FAIL with "undefined: getBrowserCommand", "undefined: buildAuthURL"

**Step 3: Implement browser opener**

Create `internal/oauth/browser.go`:

```go
package oauth

import (
	"fmt"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
)

// OpenBrowser opens the given URL in the user's default browser
// Falls back to printing the URL if unable to open browser
func OpenBrowser(authURL string) error {
	cmd, fallback := getBrowserCommand(runtime.GOOS, authURL)

	if err := cmd.Start(); err != nil {
		if fallback {
			return fmt.Errorf("could not open browser automatically")
		}
		return err
	}

	return nil
}

// getBrowserCommand returns the command to open a URL in the default browser
// Second return value indicates if this is a fallback (print URL)
func getBrowserCommand(goos, url string) (*exec.Cmd, bool) {
	switch goos {
	case "darwin":
		return exec.Command("open", url), false
	case "linux":
		return exec.Command("xdg-open", url), false
	case "windows":
		return exec.Command("cmd", "/c", "start", url), false
	default:
		// Unknown OS - return dummy command that will fail
		return exec.Command("false"), true
	}
}

// buildAuthURL constructs the OAuth authorization URL
func buildAuthURL(service, clientID, redirectURI, state string) string {
	baseURL := getServiceAuthURL(service)

	params := url.Values{}
	params.Set("client_id", clientID)
	params.Set("redirect_uri", redirectURI)
	params.Set("response_type", "code")
	params.Set("state", state)
	params.Set("scope", "tasks:read tasks:write")

	return fmt.Sprintf("%s?%s", baseURL, params.Encode())
}

// getServiceAuthURL returns the authorization URL for the given service
func getServiceAuthURL(service string) string {
	switch strings.ToLower(service) {
	case "ticktick":
		return "https://ticktick.com/oauth/authorize"
	case "dida365":
		return "https://dida365.com/oauth/authorize"
	default:
		return "https://dida365.com/oauth/authorize"
	}
}

// getServiceTokenURL returns the token URL for the given service
func getServiceTokenURL(service string) string {
	switch strings.ToLower(service) {
	case "ticktick":
		return "https://ticktick.com/oauth/token"
	case "dida365":
		return "https://dida365.com/oauth/token"
	default:
		return "https://dida365.com/oauth/token"
	}
}
```

**Step 4: Run tests to verify they pass**

```bash
go test ./internal/oauth -v
```

Expected: PASS for all tests

**Step 5: Commit**

```bash
git add internal/oauth/browser.go internal/oauth/browser_test.go
git commit -m "feat(oauth): add cross-platform browser opener with auth URL builder"
```

---

## Task 3: Implement OAuth Callback Server

**Files:**
- Create: `internal/oauth/server.go`
- Create: `internal/oauth/server_test.go`

**Step 1: Write tests for callback server**

Create `internal/oauth/server_test.go`:

```go
package oauth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCallbackHandler(t *testing.T) {
	t.Run("successful callback with code", func(t *testing.T) {
		expectedState := "test_state_123"
		expectedCode := "auth_code_456"
		resultChan := make(chan callbackResult, 1)

		handler := newCallbackHandler(expectedState, resultChan)

		req := httptest.NewRequest("GET", "/callback?code="+expectedCode+"&state="+expectedState, nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		// Check response
		if w.Code != http.StatusOK {
			t.Errorf("status code = %d, want %d", w.Code, http.StatusOK)
		}

		// Check result sent to channel
		select {
		case result := <-resultChan:
			if result.err != nil {
				t.Errorf("unexpected error: %v", result.err)
			}
			if result.code != expectedCode {
				t.Errorf("code = %s, want %s", result.code, expectedCode)
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("timeout waiting for result")
		}
	})

	t.Run("callback with error", func(t *testing.T) {
		expectedState := "test_state_123"
		resultChan := make(chan callbackResult, 1)

		handler := newCallbackHandler(expectedState, resultChan)

		req := httptest.NewRequest("GET", "/callback?error=access_denied&error_description=User+denied&state="+expectedState, nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		// Check result
		select {
		case result := <-resultChan:
			if result.err == nil {
				t.Error("expected error, got nil")
			}
			if !strings.Contains(result.err.Error(), "access_denied") {
				t.Errorf("error = %v, want to contain 'access_denied'", result.err)
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("timeout waiting for result")
		}
	})

	t.Run("callback with invalid state", func(t *testing.T) {
		expectedState := "test_state_123"
		resultChan := make(chan callbackResult, 1)

		handler := newCallbackHandler(expectedState, resultChan)

		req := httptest.NewRequest("GET", "/callback?code=auth_code&state=wrong_state", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		// Check result
		select {
		case result := <-resultChan:
			if result.err == nil {
				t.Error("expected error for invalid state, got nil")
			}
			if !strings.Contains(result.err.Error(), "state") {
				t.Errorf("error = %v, want to mention 'state'", result.err)
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("timeout waiting for result")
		}
	})
}

func TestGenerateState(t *testing.T) {
	state1 := generateState()
	state2 := generateState()

	if len(state1) != 64 {
		t.Errorf("state length = %d, want 64 (32 bytes hex)", len(state1))
	}

	if state1 == state2 {
		t.Error("generateState() should return different values")
	}

	// Check it's valid hex
	if _, err := hex.DecodeString(state1); err != nil {
		t.Errorf("state is not valid hex: %v", err)
	}
}
```

**Step 2: Run tests to verify they fail**

```bash
go test ./internal/oauth -v
```

Expected: FAIL with "undefined: newCallbackHandler", "undefined: generateState", "undefined: callbackResult"

**Step 3: Implement callback server**

Create `internal/oauth/server.go`:

```go
package oauth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"time"
)

// callbackResult contains the result from the OAuth callback
type callbackResult struct {
	code string
	err  error
}

// startCallbackServer starts an HTTP server to receive the OAuth callback
// Returns the server and a channel that will receive the result
func startCallbackServer(port int, state string) (*http.Server, chan callbackResult, error) {
	resultChan := make(chan callbackResult, 1)

	handler := newCallbackHandler(state, resultChan)
	mux := http.NewServeMux()
	mux.Handle("/callback", handler)

	addr := fmt.Sprintf("localhost:%d", port)
	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// Try to listen on the port
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start callback server on port %d: %w", port, err)
	}

	// Start server in background
	go func() {
		srv.Serve(listener)
	}()

	return srv, resultChan, nil
}

// newCallbackHandler creates an HTTP handler for the OAuth callback
func newCallbackHandler(expectedState string, resultChan chan<- callbackResult) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		// Check for error response
		if errCode := query.Get("error"); errCode != "" {
			errDesc := query.Get("error_description")
			if errDesc == "" {
				errDesc = errCode
			}

			resultChan <- callbackResult{
				err: fmt.Errorf("authorization failed: %s", errDesc),
			}

			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head><title>Authorization Failed</title></head>
<body>
	<h1>Authorization Failed</h1>
	<p>%s</p>
	<p>You can close this window.</p>
</body>
</html>`, errDesc)
			return
		}

		// Validate state
		state := query.Get("state")
		if state != expectedState {
			resultChan <- callbackResult{
				err: fmt.Errorf("invalid state parameter - possible CSRF attack"),
			}

			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head><title>Security Error</title></head>
<body>
	<h1>Security Error</h1>
	<p>Invalid state parameter. Please try again.</p>
	<p>You can close this window.</p>
</body>
</html>`)
			return
		}

		// Get authorization code
		code := query.Get("code")
		if code == "" {
			resultChan <- callbackResult{
				err: fmt.Errorf("no authorization code received"),
			}

			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head><title>Error</title></head>
<body>
	<h1>Error</h1>
	<p>No authorization code received.</p>
	<p>You can close this window.</p>
</body>
</html>`)
			return
		}

		// Success!
		resultChan <- callbackResult{
			code: code,
		}

		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head><title>Authorization Successful</title></head>
<body>
	<h1>Authorization Successful!</h1>
	<p>You have successfully authorized the Dida365 CLI.</p>
	<p>You can close this window and return to your terminal.</p>
</body>
</html>`)
	}
}

// generateState generates a random state parameter for CSRF protection
func generateState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// shutdownServer gracefully shuts down the server
func shutdownServer(srv *http.Server) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return srv.Shutdown(ctx)
}
```

**Step 4: Run tests to verify they pass**

```bash
go test ./internal/oauth -v
```

Expected: PASS for all tests

**Step 5: Commit**

```bash
git add internal/oauth/server.go internal/oauth/server_test.go
git commit -m "feat(oauth): add callback server with CSRF protection and error handling"
```

---

## Task 4: Implement OAuth Flow Orchestrator

**Files:**
- Create: `internal/oauth/oauth.go`
- Create: `internal/oauth/oauth_test.go`

**Step 1: Write tests for OAuth orchestrator**

Create `internal/oauth/oauth_test.go`:

```go
package oauth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestExchangeCodeForToken(t *testing.T) {
	t.Run("successful token exchange", func(t *testing.T) {
		// Mock token server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("method = %s, want POST", r.Method)
			}

			if err := r.ParseForm(); err != nil {
				t.Fatalf("ParseForm failed: %v", err)
			}

			if r.Form.Get("grant_type") != "authorization_code" {
				t.Error("missing grant_type=authorization_code")
			}
			if r.Form.Get("code") == "" {
				t.Error("missing code parameter")
			}

			response := TokenResponse{
				AccessToken:  "access_token_123",
				RefreshToken: "refresh_token_456",
				ExpiresIn:    7200,
				TokenType:    "Bearer",
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		result, err := exchangeCodeForToken(
			server.URL,
			"test_code",
			"test_client",
			"test_secret",
			"http://localhost:8080/callback",
		)

		if err != nil {
			t.Fatalf("exchangeCodeForToken failed: %v", err)
		}

		if result.AccessToken != "access_token_123" {
			t.Errorf("AccessToken = %s, want access_token_123", result.AccessToken)
		}
		if result.RefreshToken != "refresh_token_456" {
			t.Errorf("RefreshToken = %s, want refresh_token_456", result.RefreshToken)
		}
		if result.ExpiresIn != 7200 {
			t.Errorf("ExpiresIn = %d, want 7200", result.ExpiresIn)
		}
	})

	t.Run("token exchange error response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error":             "invalid_grant",
				"error_description": "Authorization code expired",
			})
		}))
		defer server.Close()

		_, err := exchangeCodeForToken(
			server.URL,
			"invalid_code",
			"test_client",
			"test_secret",
			"http://localhost:8080/callback",
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !strings.Contains(err.Error(), "invalid_grant") {
			t.Errorf("error = %v, want to contain 'invalid_grant'", err)
		}
	})
}

func TestRefreshAccessToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %s, want POST", r.Method)
		}

		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm failed: %v", err)
		}

		if r.Form.Get("grant_type") != "refresh_token" {
			t.Error("missing grant_type=refresh_token")
		}
		if r.Form.Get("refresh_token") == "" {
			t.Error("missing refresh_token parameter")
		}

		response := TokenResponse{
			AccessToken:  "new_access_token",
			RefreshToken: "new_refresh_token",
			ExpiresIn:    7200,
			TokenType:    "Bearer",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	result, err := refreshAccessToken(
		server.URL,
		"old_refresh_token",
		"test_client",
		"test_secret",
	)

	if err != nil {
		t.Fatalf("refreshAccessToken failed: %v", err)
	}

	if result.AccessToken != "new_access_token" {
		t.Errorf("AccessToken = %s, want new_access_token", result.AccessToken)
	}
	if result.RefreshToken != "new_refresh_token" {
		t.Errorf("RefreshToken = %s, want new_refresh_token", result.RefreshToken)
	}
}
```

**Step 2: Run tests to verify they fail**

```bash
go test ./internal/oauth -v
```

Expected: FAIL with "undefined: exchangeCodeForToken", "undefined: refreshAccessToken", "undefined: TokenResponse"

**Step 3: Implement OAuth orchestrator**

Create `internal/oauth/oauth.go`:

```go
package oauth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// TokenResponse represents the OAuth token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// StartFlow initiates the OAuth2 authorization flow
// Returns token response or error
func StartFlow(clientID, clientSecret string, port int, service string) (*TokenResponse, error) {
	// Generate state for CSRF protection
	state := generateState()

	// Build redirect URI
	redirectURI := fmt.Sprintf("http://localhost:%d/callback", port)

	// Start callback server
	srv, resultChan, err := startCallbackServer(port, state)
	if err != nil {
		return nil, err
	}
	defer shutdownServer(srv)

	// Build authorization URL
	authURL := buildAuthURL(service, clientID, redirectURI, state)

	// Open browser
	if err := OpenBrowser(authURL); err != nil {
		fmt.Printf("\nCould not open browser automatically. Please open this URL manually:\n%s\n\n", authURL)
	}

	// Wait for callback with timeout
	select {
	case result := <-resultChan:
		if result.err != nil {
			return nil, result.err
		}

		// Exchange code for token
		tokenURL := getServiceTokenURL(service)
		return exchangeCodeForToken(tokenURL, result.code, clientID, clientSecret, redirectURI)

	case <-time.After(5 * time.Minute):
		return nil, fmt.Errorf("authorization timeout - no response received within 5 minutes")
	}
}

// RefreshToken refreshes an access token using a refresh token
func RefreshToken(clientID, clientSecret, refreshToken, service string) (*TokenResponse, error) {
	tokenURL := getServiceTokenURL(service)
	return refreshAccessToken(tokenURL, refreshToken, clientID, clientSecret)
}

// exchangeCodeForToken exchanges an authorization code for tokens
func exchangeCodeForToken(tokenURL, code, clientID, clientSecret, redirectURI string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("redirect_uri", redirectURI)

	return requestToken(tokenURL, data)
}

// refreshAccessToken exchanges a refresh token for new tokens
func refreshAccessToken(tokenURL, refreshToken, clientID, clientSecret string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)

	return requestToken(tokenURL, data)
}

// requestToken makes the token request and parses the response
func requestToken(tokenURL string, data url.Values) (*TokenResponse, error) {
	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		// Try to parse error response
		var errResp struct {
			Error            string `json:"error"`
			ErrorDescription string `json:"error_description"`
		}
		if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
			return nil, fmt.Errorf("token exchange failed: %s - %s", errResp.Error, errResp.ErrorDescription)
		}
		return nil, fmt.Errorf("token exchange failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	return &tokenResp, nil
}
```

**Step 4: Run tests to verify they pass**

```bash
go test ./internal/oauth -v
```

Expected: PASS for all tests

**Step 5: Commit**

```bash
git add internal/oauth/oauth.go internal/oauth/oauth_test.go
git commit -m "feat(oauth): add OAuth2 flow orchestrator with token exchange"
```

---

## Task 5: Implement `auth login` Command

**Files:**
- Modify: `cmd/auth.go`

**Step 1: Update auth.go to add login command**

Replace the `auth configure` command section with `auth login`:

```go
// Remove old authConfigureCmd and its flags

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate via OAuth2 browser flow",
	Long:  `Opens your browser to authenticate with Dida365/TickTick via OAuth2. Automatically saves tokens with refresh capability.`,
	RunE:  runAuthLogin,
}

var (
	loginClientID     string
	loginClientSecret string
	loginService      string
	loginPort         int
)

func init() {
	// Add auth command to root
	rootCmd.AddCommand(authCmd)

	// Add subcommands to auth
	authCmd.AddCommand(authLoginCmd)  // Changed from authConfigureCmd
	authCmd.AddCommand(authStatusCmd)

	// Add flags to login command
	authLoginCmd.Flags().StringVar(&loginClientID, "client-id", "", "Dida365/TickTick API client ID (required)")
	authLoginCmd.Flags().StringVar(&loginClientSecret, "client-secret", "", "Dida365/TickTick API client secret (required)")
	authLoginCmd.Flags().StringVar(&loginService, "service", "dida365", "Service type: dida365 or ticktick (default: dida365)")
	authLoginCmd.Flags().IntVar(&loginPort, "port", 8080, "Local port for OAuth callback (default: 8080)")

	authLoginCmd.MarkFlagRequired("client-id")
	authLoginCmd.MarkFlagRequired("client-secret")
}

func runAuthLogin(cmd *cobra.Command, args []string) error {
	// Validate service
	service := strings.ToLower(loginService)
	if service != "dida365" && service != "ticktick" {
		outputError(fmt.Errorf("invalid service: %s (must be 'dida365' or 'ticktick')", loginService), "VALIDATION_ERROR", 5)
		return nil
	}

	// Start OAuth flow
	fmt.Printf("Starting OAuth2 authentication flow...\n")
	fmt.Printf("Service: %s\n", service)
	fmt.Printf("Redirect URI: http://localhost:%d/callback\n\n", loginPort)
	fmt.Printf("Opening browser for authorization...\n")

	tokenResp, err := oauth.StartFlow(loginClientID, loginClientSecret, loginPort, service)
	if err != nil {
		outputError(err, "AUTH_ERROR", 2)
		return nil
	}

	// Calculate token expiry
	expiry := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	// Determine base URL based on service
	baseURL := "https://dida365.com"
	if service == "ticktick" {
		baseURL = "https://ticktick.com"
	}

	// Create config
	cfg := &config.Config{
		ClientID:     loginClientID,
		ClientSecret: loginClientSecret,
		AccessToken:  tokenResp.AccessToken,
		BaseURL:      baseURL,
		RefreshToken: tokenResp.RefreshToken,
		TokenExpiry:  expiry,
	}

	// Save config
	configPath := config.DefaultConfigPath()
	if configPath == "" {
		outputError(fmt.Errorf("failed to determine home directory"), "CONFIG_ERROR", 1)
		return nil
	}

	if err := cfg.Save(configPath); err != nil {
		outputError(err, "SAVE_ERROR", 1)
		return nil
	}

	// Output success
	outputJSON(map[string]interface{}{
		"authenticated": true,
		"service":       service,
		"expires_at":    expiry.Format(time.RFC3339),
		"config_path":   configPath,
	})

	return nil
}
```

Add the necessary import:

```go
import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bearzk/dida365-cli/internal/client"
	"github.com/bearzk/dida365-cli/internal/config"
	"github.com/bearzk/dida365-cli/internal/oauth"
	"github.com/spf13/cobra"
)
```

**Step 2: Test the login command**

```bash
go build -o dida365 .
./dida365 auth login --help
```

Expected: Help text shows with --client-id, --client-secret, --service, --port flags

**Step 3: Commit**

```bash
git add cmd/auth.go
git commit -m "feat(auth): replace configure with login command using OAuth2 flow"
```

---

## Task 6: Implement `auth refresh` Command

**Files:**
- Modify: `cmd/auth.go`

**Step 1: Add refresh command**

Add to `cmd/auth.go`:

```go
var authRefreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Refresh the access token",
	Long:  `Use the refresh token to obtain a new access token. Requires previous authentication via 'auth login'.`,
	RunE:  runAuthRefresh,
}

func init() {
	// ... existing init code ...

	// Add refresh command
	authCmd.AddCommand(authRefreshCmd)
}

func runAuthRefresh(cmd *cobra.Command, args []string) error {
	// Load config
	configPath := config.DefaultConfigPath()
	if configPath == "" {
		outputError(fmt.Errorf("failed to determine home directory"), "CONFIG_ERROR", 1)
		return nil
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		outputError(err, "CONFIG_ERROR", 1)
		return nil
	}

	// Check if refresh token exists
	if !cfg.CanRefresh() {
		outputError(
			fmt.Errorf("no refresh token available - please run 'dida365 auth login' to re-authenticate"),
			"NO_REFRESH_TOKEN",
			1,
		)
		return nil
	}

	// Determine service from base URL
	service := "dida365"
	if strings.Contains(cfg.BaseURL, "ticktick") {
		service = "ticktick"
	}

	fmt.Printf("Refreshing access token...\n")

	// Refresh token
	tokenResp, err := oauth.RefreshToken(cfg.ClientID, cfg.ClientSecret, cfg.RefreshToken, service)
	if err != nil {
		outputError(
			fmt.Errorf("failed to refresh token - please run 'dida365 auth login' to re-authenticate: %w", err),
			"REFRESH_FAILED",
			2,
		)
		return nil
	}

	// Update config with new tokens
	expiry := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	cfg.AccessToken = tokenResp.AccessToken
	cfg.RefreshToken = tokenResp.RefreshToken
	cfg.TokenExpiry = expiry

	// Save updated config
	if err := cfg.Save(configPath); err != nil {
		outputError(err, "SAVE_ERROR", 1)
		return nil
	}

	// Output success
	outputJSON(map[string]interface{}{
		"refreshed":  true,
		"expires_at": expiry.Format(time.RFC3339),
	})

	return nil
}
```

**Step 2: Test the refresh command**

```bash
go build -o dida365 .
./dida365 auth refresh --help
```

Expected: Help text displays

```bash
./dida365 auth refresh
```

Expected: Error about no config or no refresh token

**Step 3: Commit**

```bash
git add cmd/auth.go
git commit -m "feat(auth): add refresh command for token renewal"
```

---

## Task 7: Update `auth status` Command

**Files:**
- Modify: `cmd/auth.go`

**Step 1: Update status command to show expiry and refresh info**

Update the `runAuthStatus` function in `cmd/auth.go`:

```go
func runAuthStatus(cmd *cobra.Command, args []string) error {
	// Get config path
	configPath := config.DefaultConfigPath()
	if configPath == "" {
		outputError(fmt.Errorf("failed to determine home directory"), "CONFIG_ERROR", 1)
		return nil
	}

	// Load config
	cfg, err := config.Load(configPath)
	if err != nil {
		result := map[string]interface{}{
			"configured":  false,
			"token_valid": false,
			"can_refresh": false,
			"error":       "configuration not found",
		}
		outputJSON(result)
		os.Exit(1)
		return nil
	}

	// Validate config
	if err := cfg.Validate(); err != nil {
		result := map[string]interface{}{
			"configured":  false,
			"token_valid": false,
			"can_refresh": false,
			"error":       err.Error(),
		}
		outputJSON(result)
		os.Exit(1)
		return nil
	}

	// Test token by calling ListProjects
	c := client.NewClient(cfg)
	_, err = c.ListProjects()
	tokenValid := err == nil

	// Build result
	result := map[string]interface{}{
		"configured":  true,
		"client_id":   cfg.ClientID,
		"token_valid": tokenValid,
		"can_refresh": cfg.CanRefresh(),
		"is_expired":  cfg.IsExpired(),
	}

	// Add expiry info if available
	if !cfg.TokenExpiry.IsZero() {
		result["expires_at"] = cfg.TokenExpiry.Format(time.RFC3339)
		result["expires_in_seconds"] = int(time.Until(cfg.TokenExpiry).Seconds())
	}

	if !tokenValid {
		result["error"] = fmt.Sprintf("Token validation failed: %v", err)
		if cfg.IsExpired() {
			result["suggestion"] = "Token has expired. Run 'dida365 auth refresh' to renew it."
		}
		outputJSON(result)
		os.Exit(2)
		return nil
	}

	return outputJSON(result)
}
```

**Step 2: Test the updated status command**

```bash
go build -o dida365 .
./dida365 auth status
```

Expected: Shows configured=false or shows token info with new fields (can_refresh, is_expired, expires_at, etc.)

**Step 3: Commit**

```bash
git add cmd/auth.go
git commit -m "feat(auth): enhance status command with token expiry and refresh info"
```

---

## Task 8: Update README Documentation

**Files:**
- Modify: `README.md`

**Step 1: Update authentication section**

Replace the "Getting Started" section in README.md:

```markdown
## Getting Started

### 1. Obtain API Credentials

1. Visit the developer portal:
   - For Dida365: https://developer.dida365.com/manage
   - For TickTick: https://developer.ticktick.com/manage
2. Click "New App" to create a new application
3. You'll receive your `client_id` and `client_secret`
4. Add the OAuth redirect URL: `http://localhost:8080/callback`
   - Or use a different port and specify it with `--port` flag
5. Save the changes

### 2. Authenticate via OAuth2

```bash
# For Dida365 (default)
dida365 auth login \
  --client-id "your_client_id" \
  --client-secret "your_client_secret"

# For TickTick
dida365 auth login \
  --client-id "your_client_id" \
  --client-secret "your_client_secret" \
  --service ticktick

# Custom port (if 8080 is busy)
dida365 auth login \
  --client-id "your_client_id" \
  --client-secret "your_client_secret" \
  --port 9000
```

This will:
1. Open your browser for authorization
2. Automatically exchange the authorization code for tokens
3. Save your credentials to `~/.dida365/config.json` with secure permissions

### 3. Verify Authentication

```bash
dida365 auth status
```

### 4. Refresh Token When Expired

Tokens expire after 2 hours. To refresh:

```bash
dida365 auth refresh
```
```

Update the "Exit Codes" section to add code 4:

```markdown
- `4` - Refresh error (refresh token expired or invalid)
```

Update the "Configuration File" section:

```markdown
## Configuration File

Location: `~/.dida365/config.json`

```json
{
  "client_id": "your_client_id",
  "client_secret": "your_client_secret",
  "access_token": "your_access_token",
  "base_url": "https://dida365.com",
  "refresh_token": "your_refresh_token",
  "token_expiry": "2026-02-21T16:30:00Z"
}
```

**Security:** The config file is created with `0600` permissions (user read/write only).

**Token Management:**
- Access tokens expire after 2 hours
- Use `dida365 auth refresh` to renew without re-authenticating
- Refresh tokens are long-lived but may eventually expire
```

**Step 2: Commit**

```bash
git add README.md
git commit -m "docs: update README with OAuth2 authentication flow"
```

---

## Task 9: Update CLAUDE.md

**Files:**
- Modify: `CLAUDE.md`

**Step 1: Update authentication section**

Update the "Authentication Flow (TODO)" section to "Authentication Flow":

```markdown
### Authentication Flow
✅ **OAuth2 implementation complete**

**Commands:**
- `auth login` - OAuth2 browser flow (replaces old `auth configure`)
- `auth refresh` - Refresh expired access tokens
- `auth status` - Check token status and expiry

**Flow:**
1. Start local server on `http://localhost:{port}/callback` (default port: 8080)
2. Open browser to authorization URL with state parameter (CSRF protection)
3. User authorizes in browser
4. Callback receives auth code
5. Exchange code for access_token + refresh_token
6. Save tokens with expiry to config

**Token Management:**
- Access tokens expire after 2 hours (7200 seconds)
- Refresh tokens are long-lived
- Use `auth refresh` to renew access token
- `auth status` shows expiry and refresh capability

**Services Supported:**
- Dida365 (default): `https://dida365.com`
- TickTick: `https://ticktick.com` (use `--service ticktick`)
```

**Step 2: Commit**

```bash
git add CLAUDE.md
git commit -m "docs: update CLAUDE.md - OAuth2 flow implemented"
```

---

## Task 10: Integration Testing

**Files:**
- None (testing only)

**Step 1: Build and test help commands**

```bash
go build -o dida365 .
./dida365 auth --help
./dida365 auth login --help
./dida365 auth refresh --help
./dida365 auth status
```

Expected:
- auth shows login, refresh, status subcommands (no configure)
- login shows --client-id, --client-secret, --service, --port flags
- refresh shows help text
- status shows error (no config yet)

**Step 2: Run all unit tests**

```bash
go test ./... -v
```

Expected: All tests PASS including new oauth package tests

**Step 3: Test with coverage**

```bash
go test ./... -cover
```

Expected: Good coverage for new code (aim for >80%)

**Step 4: Manual OAuth flow test (optional - requires real credentials)**

If you have Dida365/TickTick developer credentials:

```bash
# Get credentials from https://developer.dida365.com/manage
# Add redirect URL: http://localhost:8080/callback

./dida365 auth login \
  --client-id "YOUR_CLIENT_ID" \
  --client-secret "YOUR_CLIENT_SECRET" \
  --service dida365

# Browser should open, authorize, and tokens should be saved

./dida365 auth status
# Should show configured=true, token_valid=true, with expiry info

./dida365 project list
# Should work with the new token

./dida365 auth refresh
# Should successfully refresh the token
```

**Step 5: Verify no regressions**

Test that existing commands still work:

```bash
./dida365 project --help
./dida365 task --help
./dida365 --version
```

Expected: All commands still work as before

**Step 6: Final commit (if any fixes needed)**

```bash
git add .
git commit -m "test: verify OAuth2 flow integration"
```

---

## Summary

**Total Tasks:** 10
**Estimated Time:** 2-3 hours
**Lines of Code:** ~660 new lines

**What We Built:**
1. ✅ Config extended with refresh token and expiry
2. ✅ Browser opener (cross-platform)
3. ✅ OAuth callback server with CSRF protection
4. ✅ OAuth flow orchestrator with token exchange
5. ✅ `auth login` command (replaces `configure`)
6. ✅ `auth refresh` command
7. ✅ Enhanced `auth status` command
8. ✅ Updated documentation (README, CLAUDE.md)
9. ✅ Comprehensive tests
10. ✅ Integration testing

**Breaking Changes:**
- `auth configure` command removed
- Must use `auth login` for authentication

**Backward Compatibility:**
- Old configs still work (just missing refresh capability)
- `auth status` shows `can_refresh: false` for old configs
