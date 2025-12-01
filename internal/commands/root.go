package commands

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "now-sc",
	Short: "CLI tool for bootstrapping presales projects for solution consultants",
	Long: `Now-SC is a CLI tool that helps solution consultants bootstrap and manage
presales projects with structured directories, prompt templates, and AI-powered workflows.`,
	Version: "1.0.0",
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Add subcommands
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(promptCmd)
}
