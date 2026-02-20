package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
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

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
