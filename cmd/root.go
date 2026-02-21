package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	_ "github.com/spf13/viper" // Required by specification
)

var rootCmd = &cobra.Command{
	Use:   "dida365",
	Short: "CLI for Dida365 task management",
	Long:  `A command-line interface for Dida365 (滴答清单) task management, designed for automation and scripting workflows.`,
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Version = "0.1.0"
}
