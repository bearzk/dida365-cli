package cmd

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

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication configuration",
	Long:  `Configure and validate Dida365 API credentials.`,
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate via OAuth2 authorization code flow",
	Long:  `Start the OAuth2 authorization code flow to authenticate with Dida365 or TickTick. Opens a browser for user authorization and saves the tokens.`,
	RunE:  runAuthLogin,
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check authentication status",
	Long:  `Display the current authentication configuration and validate the access token.`,
	RunE:  runAuthStatus,
}

var authRefreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Refresh the access token",
	Long:  `Refresh the access token using the stored refresh token.`,
	RunE:  runAuthRefresh,
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
	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authStatusCmd)
	authCmd.AddCommand(authRefreshCmd)

	// Add flags to login command
	authLoginCmd.Flags().StringVar(&loginClientID, "client-id", "", "OAuth client ID (required)")
	authLoginCmd.Flags().StringVar(&loginClientSecret, "client-secret", "", "OAuth client secret (required)")
	authLoginCmd.Flags().StringVar(&loginService, "service", "dida365", "Service to authenticate with (dida365 or ticktick)")
	authLoginCmd.Flags().IntVar(&loginPort, "port", 8080, "Local port for OAuth callback server")

	authLoginCmd.MarkFlagRequired("client-id")
	authLoginCmd.MarkFlagRequired("client-secret")
}

func runAuthLogin(cmd *cobra.Command, args []string) error {
	// Validate service
	service := strings.ToLower(loginService)
	if service != "dida365" && service != "ticktick" {
		outputError(fmt.Errorf("invalid service: %s (must be dida365 or ticktick)", loginService), "VALIDATION_ERROR", 5)
		return nil
	}

	// Determine base URL from service
	var baseURL string
	if service == "dida365" {
		baseURL = "https://dida365.com"
	} else {
		baseURL = "https://ticktick.com"
	}

	// Print progress messages
	fmt.Fprintf(os.Stderr, "Starting OAuth2 authorization flow...\n")
	fmt.Fprintf(os.Stderr, "Service: %s\n", service)
	fmt.Fprintf(os.Stderr, "Redirect URI: http://localhost:%d/callback\n", loginPort)
	fmt.Fprintf(os.Stderr, "Opening browser for authorization...\n")

	// Start OAuth flow
	tokenResp, err := oauth.StartFlow(loginClientID, loginClientSecret, loginPort, service)
	if err != nil {
		outputError(err, "AUTH_ERROR", 2)
		return nil
	}

	// Calculate token expiry
	tokenExpiry := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	// Create config
	cfg := &config.Config{
		ClientID:     loginClientID,
		ClientSecret: loginClientSecret,
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		TokenExpiry:  tokenExpiry,
		BaseURL:      baseURL,
	}

	// Save config
	configPath := config.DefaultConfigPath()
	if configPath == "" {
		outputError(fmt.Errorf("failed to determine home directory"), "CONFIG_ERROR", 1)
		return nil
	}

	if err := cfg.Save(configPath); err != nil {
		outputError(err, "CONFIG_ERROR", 1)
		return nil
	}

	// Output success
	outputJSON(map[string]interface{}{
		"authenticated": true,
		"service":       service,
		"expires_at":    tokenExpiry.Format(time.RFC3339),
		"config_path":   configPath,
	})

	return nil
}

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
		outputError(err, "CONFIG_NOT_FOUND", 1)
		return nil
	}

	// Validate config
	if err := cfg.Validate(); err != nil {
		outputError(err, "VALIDATION_ERROR", 1)
		return nil
	}

	// Test token by calling ListProjects
	c := client.NewClient(cfg)
	_, err = c.ListProjects()
	if err != nil {
		// Token is invalid
		result := map[string]interface{}{
			"configured":  true,
			"client_id":   cfg.ClientID,
			"token_valid": false,
			"error":       fmt.Sprintf("Token validation failed: %v", err),
			"can_refresh": cfg.CanRefresh(),
			"is_expired":  cfg.IsExpired(),
		}
		// Add expiry fields if available
		if !cfg.TokenExpiry.IsZero() {
			result["expires_at"] = cfg.TokenExpiry.Format(time.RFC3339)
			result["expires_in_seconds"] = int(time.Until(cfg.TokenExpiry).Seconds())
		}
		// Add suggestion if token is expired
		if cfg.IsExpired() {
			result["suggestion"] = "Token has expired. Run 'dida365 auth refresh' to renew it."
		}
		outputJSON(result)
		os.Exit(2)
		return nil
	}

	// Token is valid
	result := map[string]interface{}{
		"configured":  true,
		"client_id":   cfg.ClientID,
		"token_valid": true,
		"can_refresh": cfg.CanRefresh(),
		"is_expired":  cfg.IsExpired(),
	}
	// Add expiry fields if available
	if !cfg.TokenExpiry.IsZero() {
		result["expires_at"] = cfg.TokenExpiry.Format(time.RFC3339)
		result["expires_in_seconds"] = int(time.Until(cfg.TokenExpiry).Seconds())
	}
	// Add suggestion if token is expired
	if cfg.IsExpired() {
		result["suggestion"] = "Token has expired. Run 'dida365 auth refresh' to renew it."
	}
	outputJSON(result)

	return nil
}

func runAuthRefresh(cmd *cobra.Command, args []string) error {
	// Get config path
	configPath := config.DefaultConfigPath()
	if configPath == "" {
		outputError(fmt.Errorf("failed to determine home directory"), "CONFIG_ERROR", 1)
		return nil
	}

	// Load config
	cfg, err := config.Load(configPath)
	if err != nil {
		outputError(err, "CONFIG_ERROR", 1)
		return nil
	}

	// Check if refresh token exists
	if !cfg.CanRefresh() {
		outputError(fmt.Errorf("no refresh token available"), "NO_REFRESH_TOKEN", 1)
		return nil
	}

	// Determine service from BaseURL
	var service string
	if strings.Contains(cfg.BaseURL, "ticktick") {
		service = "ticktick"
	} else {
		service = "dida365"
	}

	// Print progress message
	fmt.Fprintf(os.Stderr, "Refreshing access token...\n")

	// Call RefreshToken
	tokenResp, err := oauth.RefreshToken(cfg.ClientID, cfg.ClientSecret, cfg.RefreshToken, service)
	if err != nil {
		outputError(err, "REFRESH_FAILED", 2)
		return nil
	}

	// Calculate new expiry
	tokenExpiry := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	// Update config with new tokens
	cfg.AccessToken = tokenResp.AccessToken
	cfg.RefreshToken = tokenResp.RefreshToken
	cfg.TokenExpiry = tokenExpiry

	// Save config
	if err := cfg.Save(configPath); err != nil {
		outputError(err, "CONFIG_ERROR", 1)
		return nil
	}

	// Output success
	outputJSON(map[string]interface{}{
		"refreshed":  true,
		"expires_at": tokenExpiry.Format(time.RFC3339),
	})

	return nil
}

// outputJSON outputs data as JSON to stdout
func outputJSON(data interface{}) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(jsonData))
}

// outputError outputs an error message as JSON to stderr and exits with the given code
func outputError(err error, code string, exitCode int) error {
	errObj := map[string]interface{}{
		"error": err.Error(),
		"code":  code,
	}
	encoder := json.NewEncoder(os.Stderr)
	encoder.SetIndent("", "  ")
	encoder.Encode(errObj)
	os.Exit(exitCode)
	return nil
}
