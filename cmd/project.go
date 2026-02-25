package cmd

import (
	"fmt"

	"github.com/bearzk/dida365-cli/internal/client"
	"github.com/bearzk/dida365-cli/internal/config"
	"github.com/spf13/cobra"
)

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage projects",
	Long:  `List and view Dida365 projects.`,
}

var projectListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all projects",
	Long:  `Retrieve and display all projects for the authenticated user.`,
	RunE:  runProjectList,
}

var projectGetCmd = &cobra.Command{
	Use:   "get <project-id>",
	Short: "Get project details",
	Long:  `Retrieve detailed information about a specific project.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runProjectGet,
}

var projectDataCmd = &cobra.Command{
	Use:   "data <project-id>",
	Short: "Print raw project data response",
	Long:  `Print the raw JSON response from GET /open/v1/project/{id}/data. Use this to inspect the full response shape including columns and other fields.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runProjectData,
}

func init() {
	// Add project command to root
	rootCmd.AddCommand(projectCmd)

	// Add subcommands to project
	projectCmd.AddCommand(projectListCmd)
	projectCmd.AddCommand(projectGetCmd)
	projectCmd.AddCommand(projectDataCmd)
}

func runProjectList(cmd *cobra.Command, args []string) error {
	c := loadClient()

	projects, err := c.ListProjects()
	if err != nil {
		outputError(err, "API_ERROR", 3)
		return nil
	}

	outputJSON(projects)
	return nil
}

func runProjectGet(cmd *cobra.Command, args []string) error {
	projectID := args[0]
	c := loadClient()

	project, err := c.GetProject(projectID)
	if err != nil {
		outputError(err, "API_ERROR", 3)
		return nil
	}

	outputJSON(project)
	return nil
}

func runProjectData(cmd *cobra.Command, args []string) error {
	projectID := args[0]
	c := loadClient()

	data, err := c.GetProjectData(projectID)
	if err != nil {
		outputError(err, "API_ERROR", 3)
		return nil
	}

	outputJSON(data)
	return nil
}

// loadClient loads the config and creates a new API client
// Exits with code 1 if config cannot be loaded or validated
func loadClient() *client.Client {
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

	// Validate config
	if err := cfg.Validate(); err != nil {
		outputError(err, "CONFIG_ERROR", 1)
		return nil
	}

	return client.NewClient(cfg)
}
