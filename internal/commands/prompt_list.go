package commands

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var promptListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available prompt templates",
	Long:  `Lists all prompt templates available in the 10_PromptTemplates directory.`,
	RunE:  runPromptList,
}

func runPromptList(cmd *cobra.Command, args []string) error {
	// Get current directory as project root
	projectRoot := "."

	// List all prompts
	prompts, err := ListPrompts(projectRoot)
	if err != nil {
		color.Red("Error: %v", err)
		return err
	}

	// Display prompts
	fmt.Println()
	color.Cyan("Available Prompts:")
	fmt.Println()

	for i, prompt := range prompts {
		fmt.Printf("%2d. ", i+1)
		color.Green(prompt.Name)
		if prompt.Description != "" {
			fmt.Printf("    %s\n", color.New(color.Faint).Sprint(prompt.Description))
		} else {
			fmt.Println()
		}
	}

	fmt.Println()
	color.Yellow("Usage:")
	fmt.Println("  Run a prompt: now-sc prompt run <name>")
	fmt.Println("  Interactive:  now-sc prompt")

	return nil
}
