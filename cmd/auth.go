package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/bearzk/dida365-cli/internal/client"
	"github.com/bearzk/dida365-cli/internal/config"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication configuration",
	Long:  `Configure and validate Dida365 API credentials.`,
}

var authConfigureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure authentication credentials",
	Long:  `Configure the CLI with your Dida365 API credentials. Tests the connection before saving.`,
	RunE:  runAuthConfigure,
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check authentication status",
	Long:  `Display the current authentication configuration and validate the access token.`,
	RunE:  runAuthStatus,
}

var (
	clientID     string
	clientSecret string
	accessToken  string
)

func init() {
	// Add auth command to root
	rootCmd.AddCommand(authCmd)

	// Add subcommands to auth
	authCmd.AddCommand(authConfigureCmd)
	authCmd.AddCommand(authStatusCmd)

	// Add flags to configure command
	authConfigureCmd.Flags().StringVar(&clientID, "client-id", "", "Dida365 API client ID (required)")
	authConfigureCmd.Flags().StringVar(&clientSecret, "client-secret", "", "Dida365 API client secret (required)")
	authConfigureCmd.Flags().StringVar(&accessToken, "access-token", "", "Dida365 API access token (required)")

	authConfigureCmd.MarkFlagRequired("client-id")
	authConfigureCmd.MarkFlagRequired("client-secret")
	authConfigureCmd.MarkFlagRequired("access-token")
}

func runAuthConfigure(cmd *cobra.Command, args []string) error {
	// Create config
	cfg := &config.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		AccessToken:  accessToken,
		BaseURL:      "https://dida365.com",
	}

	// Validate config
	if err := cfg.Validate(); err != nil {
		outputError(err, "VALIDATION_ERROR", 5)
		return nil
	}

	// Test connection by calling ListProjects
	c := client.NewClient(cfg)
	if _, err := c.ListProjects(); err != nil {
		outputError(err, "CONNECTION_ERROR", 2)
		return nil
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
		"configured":  true,
		"config_path": configPath,
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
		}
		outputJSON(result)
		os.Exit(2)
		return nil
	}

	// Token is valid
	outputJSON(map[string]interface{}{
		"configured":  true,
		"client_id":   cfg.ClientID,
		"token_valid": true,
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
