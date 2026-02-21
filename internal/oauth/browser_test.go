package oauth

import (
	"net/url"
	"strings"
	"testing"
)

func TestGetBrowserCommand(t *testing.T) {
	tests := []struct {
		name           string
		goos           string
		url            string
		wantCmd        string
		wantArgs       []string
		wantFallback   bool
	}{
		{
			name:         "darwin uses open",
			goos:         "darwin",
			url:          "https://example.com",
			wantCmd:      "open",
			wantArgs:     []string{"https://example.com"},
			wantFallback: false,
		},
		{
			name:         "linux uses xdg-open",
			goos:         "linux",
			url:          "https://example.com",
			wantCmd:      "xdg-open",
			wantArgs:     []string{"https://example.com"},
			wantFallback: false,
		},
		{
			name:         "windows uses cmd /c start",
			goos:         "windows",
			url:          "https://example.com",
			wantCmd:      "cmd",
			wantArgs:     []string{"/c", "start", "https://example.com"},
			wantFallback: false,
		},
		{
			name:         "unknown OS returns fallback",
			goos:         "unknown",
			url:          "https://example.com",
			wantCmd:      "",
			wantArgs:     nil,
			wantFallback: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, fallback := getBrowserCommand(tt.goos, tt.url)

			if fallback != tt.wantFallback {
				t.Errorf("getBrowserCommand() fallback = %v, want %v", fallback, tt.wantFallback)
			}

			if tt.wantFallback {
				// For fallback case, we don't care about the command details
				return
			}

			// Check that the path ends with the expected command (to handle full paths like /usr/bin/open)
			if !strings.HasSuffix(cmd.Path, tt.wantCmd) {
				t.Errorf("getBrowserCommand() cmd.Path = %v, want to end with %v", cmd.Path, tt.wantCmd)
			}

			if len(cmd.Args) != len(tt.wantArgs)+1 {
				t.Errorf("getBrowserCommand() args length = %v, want %v", len(cmd.Args), len(tt.wantArgs)+1)
				return
			}

			// cmd.Args[0] is the command itself, compare from index 1
			for i, arg := range tt.wantArgs {
				if cmd.Args[i+1] != arg {
					t.Errorf("getBrowserCommand() args[%d] = %v, want %v", i+1, cmd.Args[i+1], arg)
				}
			}
		})
	}
}

func TestBuildAuthURL(t *testing.T) {
	tests := []struct {
		name        string
		service     string
		clientID    string
		redirectURI string
		state       string
		wantHost    string
		wantParams  map[string]string
	}{
		{
			name:        "dida365 service",
			service:     "dida365",
			clientID:    "test-client-id",
			redirectURI: "http://localhost:8080/callback",
			state:       "random-state-string",
			wantHost:    "dida365.com",
			wantParams: map[string]string{
				"client_id":     "test-client-id",
				"redirect_uri":  "http://localhost:8080/callback",
				"response_type": "code",
				"state":         "random-state-string",
				"scope":         "tasks:read tasks:write",
			},
		},
		{
			name:        "ticktick service",
			service:     "ticktick",
			clientID:    "another-client-id",
			redirectURI: "http://localhost:9090/auth",
			state:       "another-state",
			wantHost:    "ticktick.com",
			wantParams: map[string]string{
				"client_id":     "another-client-id",
				"redirect_uri":  "http://localhost:9090/auth",
				"response_type": "code",
				"state":         "another-state",
				"scope":         "tasks:read tasks:write",
			},
		},
		{
			name:        "case insensitive service name",
			service:     "DIDA365",
			clientID:    "test-id",
			redirectURI: "http://localhost:8080/callback",
			state:       "state123",
			wantHost:    "dida365.com",
			wantParams: map[string]string{
				"client_id":     "test-id",
				"redirect_uri":  "http://localhost:8080/callback",
				"response_type": "code",
				"state":         "state123",
				"scope":         "tasks:read tasks:write",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authURL := buildAuthURL(tt.service, tt.clientID, tt.redirectURI, tt.state)

			// Verify the URL contains the correct host
			if !strings.Contains(authURL, tt.wantHost) {
				t.Errorf("buildAuthURL() URL = %v, want to contain host %v", authURL, tt.wantHost)
			}

			// Verify the URL contains /oauth/authorize path
			if !strings.Contains(authURL, "/oauth/authorize") {
				t.Errorf("buildAuthURL() URL = %v, want to contain path /oauth/authorize", authURL)
			}

			// Verify all required parameters are present by parsing the URL
			parsedURL := authURL
			parts := strings.Split(parsedURL, "?")
			if len(parts) != 2 {
				t.Errorf("buildAuthURL() URL = %v, expected to have query parameters", authURL)
				return
			}

			queryString := parts[1]
			params := strings.Split(queryString, "&")
			paramMap := make(map[string]string)
			for _, param := range params {
				kv := strings.SplitN(param, "=", 2)
				if len(kv) == 2 {
					// URL decode the value for comparison
					decoded, err := url.QueryUnescape(kv[1])
					if err != nil {
						t.Errorf("Failed to decode parameter value: %v", err)
						continue
					}
					paramMap[kv[0]] = decoded
				}
			}

			// Check all expected parameters
			for key, expectedValue := range tt.wantParams {
				gotValue, exists := paramMap[key]
				if !exists {
					t.Errorf("buildAuthURL() missing parameter %v", key)
					continue
				}
				if gotValue != expectedValue {
					t.Errorf("buildAuthURL() parameter %v = %v, want %v", key, gotValue, expectedValue)
				}
			}
		})
	}
}

func TestGetServiceAuthURL(t *testing.T) {
	tests := []struct {
		name    string
		service string
		want    string
	}{
		{
			name:    "dida365 auth URL",
			service: "dida365",
			want:    "https://dida365.com/oauth/authorize",
		},
		{
			name:    "ticktick auth URL",
			service: "ticktick",
			want:    "https://ticktick.com/oauth/authorize",
		},
		{
			name:    "default to dida365",
			service: "unknown",
			want:    "https://dida365.com/oauth/authorize",
		},
		{
			name:    "case insensitive",
			service: "TICKTICK",
			want:    "https://ticktick.com/oauth/authorize",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getServiceAuthURL(tt.service)
			if got != tt.want {
				t.Errorf("getServiceAuthURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetServiceTokenURL(t *testing.T) {
	tests := []struct {
		name    string
		service string
		want    string
	}{
		{
			name:    "dida365 token URL",
			service: "dida365",
			want:    "https://dida365.com/oauth/token",
		},
		{
			name:    "ticktick token URL",
			service: "ticktick",
			want:    "https://ticktick.com/oauth/token",
		},
		{
			name:    "default to dida365",
			service: "unknown",
			want:    "https://dida365.com/oauth/token",
		},
		{
			name:    "case insensitive",
			service: "TICKTICK",
			want:    "https://ticktick.com/oauth/token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getServiceTokenURL(tt.service)
			if got != tt.want {
				t.Errorf("getServiceTokenURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
