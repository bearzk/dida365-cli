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

// TokenResponse represents the OAuth token response from the authorization server.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// errorResponse represents an OAuth error response.
type errorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// StartFlow initiates and completes the OAuth 2.0 authorization code flow.
// It starts a local callback server, opens the browser for user authorization,
// waits for the callback, and exchanges the authorization code for tokens.
//
// Parameters:
//   - clientID: OAuth client identifier
//   - clientSecret: OAuth client secret
//   - port: Local port for the callback server
//   - service: Service name ("dida365" or "ticktick")
//
// Returns the token response or an error if the flow fails.
func StartFlow(clientID, clientSecret string, port int, service string) (*TokenResponse, error) {
	// Generate state for CSRF protection
	state, err := generateState()
	if err != nil {
		return nil, fmt.Errorf("failed to generate state: %w", err)
	}

	// Build redirect URI
	redirectURI := fmt.Sprintf("http://localhost:%d/callback", port)

	// Start callback server
	server, resultChan, err := startCallbackServer(port, state)
	if err != nil {
		return nil, fmt.Errorf("failed to start callback server: %w", err)
	}
	defer shutdownServer(server)

	// Build authorization URL
	authURL := buildAuthURL(service, clientID, redirectURI, state)

	// Open browser
	if err := OpenBrowser(authURL); err != nil {
		// Error is already logged by OpenBrowser, but we continue waiting for callback
		// in case the user manually opens the URL
	}

	// Wait for callback with 5 minute timeout
	var code string
	select {
	case result := <-resultChan:
		if result.err != nil {
			return nil, fmt.Errorf("callback error: %w", result.err)
		}
		code = result.code
	case <-time.After(5 * time.Minute):
		return nil, fmt.Errorf("timeout waiting for OAuth callback")
	}

	// Exchange code for token
	tokenURL := getServiceTokenURL(service)
	tokenResp, err := exchangeCodeForToken(tokenURL, code, clientID, clientSecret, redirectURI)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	return tokenResp, nil
}

// RefreshToken refreshes an access token using a refresh token.
//
// Parameters:
//   - clientID: OAuth client identifier
//   - clientSecret: OAuth client secret
//   - refreshToken: The refresh token obtained from a previous authorization
//   - service: Service name ("dida365" or "ticktick")
//
// Returns a new token response or an error if the refresh fails.
func RefreshToken(clientID, clientSecret, refreshToken, service string) (*TokenResponse, error) {
	tokenURL := getServiceTokenURL(service)
	tokenResp, err := refreshAccessToken(tokenURL, refreshToken, clientID, clientSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}
	return tokenResp, nil
}

// exchangeCodeForToken exchanges an authorization code for an access token.
//
// Parameters:
//   - tokenURL: The token endpoint URL
//   - code: The authorization code from the callback
//   - clientID: OAuth client identifier
//   - clientSecret: OAuth client secret
//   - redirectURI: The redirect URI used in the authorization request
//
// Returns the token response or an error if the exchange fails.
func exchangeCodeForToken(tokenURL, code, clientID, clientSecret, redirectURI string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("redirect_uri", redirectURI)

	return requestToken(tokenURL, data)
}

// refreshAccessToken refreshes an access token using a refresh token.
//
// Parameters:
//   - tokenURL: The token endpoint URL
//   - refreshToken: The refresh token
//   - clientID: OAuth client identifier
//   - clientSecret: OAuth client secret
//
// Returns the token response or an error if the refresh fails.
func refreshAccessToken(tokenURL, refreshToken, clientID, clientSecret string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)

	return requestToken(tokenURL, data)
}

// requestToken performs a token request to the OAuth server.
// This is the shared HTTP logic for both authorization code exchange and token refresh.
//
// Parameters:
//   - tokenURL: The token endpoint URL
//   - data: Form data to send in the request body
//
// Returns the token response or an error if the request fails.
func requestToken(tokenURL string, data url.Values) (*TokenResponse, error) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Create request
	req, err := http.NewRequest(http.MethodPost, tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for error response
	if resp.StatusCode != http.StatusOK {
		var errResp errorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			// If we can't parse the error response, return a generic error
			return nil, fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
		}
		return nil, fmt.Errorf("OAuth error: %s - %s", errResp.Error, errResp.ErrorDescription)
	}

	// Parse success response
	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	return &tokenResp, nil
}
