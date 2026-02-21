package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestConfigLoad(t *testing.T) {
	t.Run("load valid config", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.json")

		validConfig := Config{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			AccessToken:  "test-access-token",
			BaseURL:      "https://api.dida365.com",
		}

		data, err := json.Marshal(validConfig)
		if err != nil {
			t.Fatalf("failed to marshal config: %v", err)
		}

		if err := os.WriteFile(configPath, data, 0600); err != nil {
			t.Fatalf("failed to write config file: %v", err)
		}

		loaded, err := Load(configPath)
		if err != nil {
			t.Fatalf("Load() error = %v, want nil", err)
		}

		if loaded.ClientID != validConfig.ClientID {
			t.Errorf("ClientID = %s, want %s", loaded.ClientID, validConfig.ClientID)
		}
		if loaded.ClientSecret != validConfig.ClientSecret {
			t.Errorf("ClientSecret = %s, want %s", loaded.ClientSecret, validConfig.ClientSecret)
		}
		if loaded.AccessToken != validConfig.AccessToken {
			t.Errorf("AccessToken = %s, want %s", loaded.AccessToken, validConfig.AccessToken)
		}
		if loaded.BaseURL != validConfig.BaseURL {
			t.Errorf("BaseURL = %s, want %s", loaded.BaseURL, validConfig.BaseURL)
		}
	})

	t.Run("load nonexistent config", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "nonexistent.json")

		_, err := Load(configPath)
		if err == nil {
			t.Error("Load() error = nil, want error for nonexistent file")
		}
	})

	t.Run("load invalid JSON", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.json")

		invalidJSON := []byte(`{"clientID": "test", "invalid`)
		if err := os.WriteFile(configPath, invalidJSON, 0600); err != nil {
			t.Fatalf("failed to write config file: %v", err)
		}

		_, err := Load(configPath)
		if err == nil {
			t.Error("Load() error = nil, want error for invalid JSON")
		}
	})
}

func TestConfigSave(t *testing.T) {
	t.Run("save valid config", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.json")

		config := Config{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			AccessToken:  "test-access-token",
			BaseURL:      "https://api.dida365.com",
		}

		err := config.Save(configPath)
		if err != nil {
			t.Fatalf("Save() error = %v, want nil", err)
		}

		// Verify file exists
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("config file was not created")
		}

		// Verify file permissions are 0600
		info, err := os.Stat(configPath)
		if err != nil {
			t.Fatalf("failed to stat config file: %v", err)
		}
		if info.Mode().Perm() != 0600 {
			t.Errorf("file permissions = %o, want 0600", info.Mode().Perm())
		}

		// Verify content
		loaded, err := Load(configPath)
		if err != nil {
			t.Fatalf("failed to load saved config: %v", err)
		}
		if loaded.ClientID != config.ClientID {
			t.Errorf("ClientID = %s, want %s", loaded.ClientID, config.ClientID)
		}
		if loaded.ClientSecret != config.ClientSecret {
			t.Errorf("ClientSecret = %s, want %s", loaded.ClientSecret, config.ClientSecret)
		}
		if loaded.AccessToken != config.AccessToken {
			t.Errorf("AccessToken = %s, want %s", loaded.AccessToken, config.AccessToken)
		}
		if loaded.BaseURL != config.BaseURL {
			t.Errorf("BaseURL = %s, want %s", loaded.BaseURL, config.BaseURL)
		}
	})

	t.Run("save creates directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		nestedPath := filepath.Join(tmpDir, "nested", "dir", "config.json")

		config := Config{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			AccessToken:  "test-access-token",
			BaseURL:      "https://api.dida365.com",
		}

		err := config.Save(nestedPath)
		if err != nil {
			t.Fatalf("Save() error = %v, want nil", err)
		}

		// Verify directory was created
		dirPath := filepath.Dir(nestedPath)
		info, err := os.Stat(dirPath)
		if err != nil {
			t.Fatalf("directory was not created: %v", err)
		}
		if !info.IsDir() {
			t.Error("path is not a directory")
		}

		// Verify directory permissions are 0700
		if info.Mode().Perm() != 0700 {
			t.Errorf("directory permissions = %o, want 0700", info.Mode().Perm())
		}

		// Verify file was created
		if _, err := os.Stat(nestedPath); os.IsNotExist(err) {
			t.Error("config file was not created")
		}
	})
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: Config{
				ClientID:     "test-client-id",
				ClientSecret: "test-client-secret",
				AccessToken:  "test-access-token",
				BaseURL:      "https://api.dida365.com",
			},
			wantErr: false,
		},
		{
			name: "missing ClientID",
			config: Config{
				ClientSecret: "test-client-secret",
				AccessToken:  "test-access-token",
				BaseURL:      "https://api.dida365.com",
			},
			wantErr: true,
			errMsg:  "ClientID",
		},
		{
			name: "missing ClientSecret",
			config: Config{
				ClientID:    "test-client-id",
				AccessToken: "test-access-token",
				BaseURL:     "https://api.dida365.com",
			},
			wantErr: true,
			errMsg:  "ClientSecret",
		},
		{
			name: "missing AccessToken",
			config: Config{
				ClientID:     "test-client-id",
				ClientSecret: "test-client-secret",
				BaseURL:      "https://api.dida365.com",
			},
			wantErr: true,
			errMsg:  "AccessToken",
		},
		{
			name: "missing BaseURL",
			config: Config{
				ClientID:     "test-client-id",
				ClientSecret: "test-client-secret",
				AccessToken:  "test-access-token",
			},
			wantErr: true,
			errMsg:  "BaseURL",
		},
		{
			name:    "empty config",
			config:  Config{},
			wantErr: true,
			errMsg:  "ClientID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				if err == nil {
					t.Error("Validate() error = nil, want error")
				} else if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("Validate() error = %v, want error containing %q", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() error = %v, want nil", err)
				}
			}
		})
	}
}

func TestDefaultConfigPath(t *testing.T) {
	path := DefaultConfigPath()

	// Should return a non-empty path when home directory is available
	// If home is not available, it should return empty string
	// We can't easily test the failure case, so we just check that it returns something reasonable
	if path != "" {
		expected := filepath.Join(os.Getenv("HOME"), ".dida365", "config.json")
		if path != expected {
			t.Errorf("DefaultConfigPath() = %s, want %s", path, expected)
		}
	}
}

func TestConfigWithTokenFields(t *testing.T) {
	t.Run("marshal config with refresh token and expiry", func(t *testing.T) {
		expiry := time.Now().Add(1 * time.Hour)
		config := Config{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			AccessToken:  "test-access-token",
			RefreshToken: "test-refresh-token",
			TokenExpiry:  expiry,
			BaseURL:      "https://api.dida365.com",
		}

		data, err := json.Marshal(config)
		if err != nil {
			t.Fatalf("Marshal() error = %v, want nil", err)
		}

		var unmarshaled map[string]interface{}
		if err := json.Unmarshal(data, &unmarshaled); err != nil {
			t.Fatalf("Unmarshal() error = %v, want nil", err)
		}

		if unmarshaled["refresh_token"] != "test-refresh-token" {
			t.Errorf("refresh_token = %v, want %v", unmarshaled["refresh_token"], "test-refresh-token")
		}

		if _, ok := unmarshaled["token_expiry"]; !ok {
			t.Error("token_expiry field missing in JSON")
		}
	})

	t.Run("unmarshal config with refresh token and expiry", func(t *testing.T) {
		expiry := time.Now().Add(1 * time.Hour).UTC()
		jsonData := `{
			"client_id": "test-client-id",
			"client_secret": "test-client-secret",
			"access_token": "test-access-token",
			"refresh_token": "test-refresh-token",
			"token_expiry": "` + expiry.Format(time.RFC3339) + `",
			"base_url": "https://api.dida365.com"
		}`

		var config Config
		if err := json.Unmarshal([]byte(jsonData), &config); err != nil {
			t.Fatalf("Unmarshal() error = %v, want nil", err)
		}

		if config.RefreshToken != "test-refresh-token" {
			t.Errorf("RefreshToken = %v, want %v", config.RefreshToken, "test-refresh-token")
		}

		if config.TokenExpiry.IsZero() {
			t.Error("TokenExpiry is zero, want non-zero value")
		}

		// Compare with 1 second tolerance for time parsing
		if diff := config.TokenExpiry.Sub(expiry); diff > time.Second || diff < -time.Second {
			t.Errorf("TokenExpiry = %v, want %v", config.TokenExpiry, expiry)
		}
	})

	t.Run("backward compatibility - old config without new fields", func(t *testing.T) {
		jsonData := `{
			"client_id": "test-client-id",
			"client_secret": "test-client-secret",
			"access_token": "test-access-token",
			"base_url": "https://api.dida365.com"
		}`

		var config Config
		if err := json.Unmarshal([]byte(jsonData), &config); err != nil {
			t.Fatalf("Unmarshal() error = %v, want nil", err)
		}

		if config.RefreshToken != "" {
			t.Errorf("RefreshToken = %v, want empty string", config.RefreshToken)
		}

		if !config.TokenExpiry.IsZero() {
			t.Errorf("TokenExpiry = %v, want zero time", config.TokenExpiry)
		}

		if config.ClientID != "test-client-id" {
			t.Errorf("ClientID = %v, want %v", config.ClientID, "test-client-id")
		}
	})
}

func TestConfigIsExpired(t *testing.T) {
	t.Run("no expiry set returns false", func(t *testing.T) {
		config := Config{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			AccessToken:  "test-access-token",
			BaseURL:      "https://api.dida365.com",
		}

		if config.IsExpired() {
			t.Error("IsExpired() = true, want false when no expiry set")
		}
	})

	t.Run("expired token returns true", func(t *testing.T) {
		config := Config{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			AccessToken:  "test-access-token",
			TokenExpiry:  time.Now().Add(-1 * time.Hour),
			BaseURL:      "https://api.dida365.com",
		}

		if !config.IsExpired() {
			t.Error("IsExpired() = false, want true for expired token")
		}
	})

	t.Run("valid token returns false", func(t *testing.T) {
		config := Config{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			AccessToken:  "test-access-token",
			TokenExpiry:  time.Now().Add(1 * time.Hour),
			BaseURL:      "https://api.dida365.com",
		}

		if config.IsExpired() {
			t.Error("IsExpired() = true, want false for valid token")
		}
	})

	t.Run("token expired exactly now returns true", func(t *testing.T) {
		config := Config{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			AccessToken:  "test-access-token",
			TokenExpiry:  time.Now().Add(-1 * time.Second),
			BaseURL:      "https://api.dida365.com",
		}

		if !config.IsExpired() {
			t.Error("IsExpired() = false, want true for token expired at current time")
		}
	})
}

func TestConfigCanRefresh(t *testing.T) {
	t.Run("has refresh token returns true", func(t *testing.T) {
		config := Config{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			AccessToken:  "test-access-token",
			RefreshToken: "test-refresh-token",
			BaseURL:      "https://api.dida365.com",
		}

		if !config.CanRefresh() {
			t.Error("CanRefresh() = false, want true when refresh token is set")
		}
	})

	t.Run("no refresh token returns false", func(t *testing.T) {
		config := Config{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			AccessToken:  "test-access-token",
			BaseURL:      "https://api.dida365.com",
		}

		if config.CanRefresh() {
			t.Error("CanRefresh() = true, want false when refresh token is not set")
		}
	})
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
