package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Now-AI-Foundry/Now-SC/internal/claude"
	"github.com/Now-AI-Foundry/Now-SC/internal/openrouter"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var promptCmd = &cobra.Command{
	Use:   "prompt",
	Short: "Work with prompt templates",
	Long: `Execute prompt templates interactively or via direct commands.
Supports Claude Code integration and OpenRouter API.

Subcommands:
  list - List all available prompts
  run  - Execute a specific prompt by name

Interactive mode (default):
  now-sc prompt

Direct execution:
  now-sc prompt run <name>
  cat file.txt | now-sc prompt run <name>`,
	RunE: runPrompt,
}

func init() {
	// Add subcommands
	promptCmd.AddCommand(promptListCmd)
	promptCmd.AddCommand(promptRunCmd)
}

func runPrompt(cmd *cobra.Command, args []string) error {
	// Check for Claude Code or API key
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	hasClaudeCode := claude.IsAvailable()

	if apiKey == "" && !hasClaudeCode {
		color.Red("Error: No AI provider available")
		color.Yellow("Please either:")
		fmt.Println("  1. Install Claude Code: https://claude.ai/download")
		fmt.Println("  2. Set OPENROUTER_API_KEY environment variable")
		return fmt.Errorf("no AI provider configured")
	}

	// Determine which provider to use
	useClaudeCode := hasClaudeCode
	if apiKey != "" && !hasClaudeCode {
		useClaudeCode = false
	}

	// Find prompt templates directory
	promptsPath := filepath.Join(".", "10_PromptTemplates")
	if _, err := os.Stat(promptsPath); os.IsNotExist(err) {
		color.Red("Error: No prompt templates directory found in current directory")
		color.Yellow("Make sure you are in a project created with \"now-sc init\"")
		return fmt.Errorf("prompt templates directory not found")
	}

	// List available prompts
	entries, err := os.ReadDir(promptsPath)
	if err != nil {
		return fmt.Errorf("failed to read prompts directory: %w", err)
	}

	var promptFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			promptFiles = append(promptFiles, entry.Name())
		}
	}

	if len(promptFiles) == 0 {
		color.Red("Error: No prompt templates found")
		return fmt.Errorf("no prompt templates found")
	}

	// Let user select a prompt
	templates := make([]string, len(promptFiles))
	for i, file := range promptFiles {
		templates[i] = strings.TrimSuffix(strings.ReplaceAll(file, "_", " "), ".md")
	}

	promptSelect := promptui.Select{
		Label: "Select a prompt template",
		Items: templates,
	}

	idx, _, err := promptSelect.Run()
	if err != nil {
		return fmt.Errorf("prompt selection failed: %w", err)
	}

	selectedPrompt := promptFiles[idx]

	// Read the prompt content
	promptContent, err := os.ReadFile(filepath.Join(promptsPath, selectedPrompt))
	if err != nil {
		return fmt.Errorf("failed to read prompt file: %w", err)
	}

	// Show prompt preview
	fmt.Println()
	color.Cyan("Prompt Preview:")
	fmt.Println("─────────────────────────────────────────")
	preview := string(promptContent)
	if len(preview) > 200 {
		preview = preview[:200] + "..."
	}
	fmt.Println(preview)
	fmt.Println("─────────────────────────────────────────")

	// Get user input
	promptInput := promptui.Prompt{
		Label: "Enter your input for this prompt",
	}
	userInput, err := promptInput.Run()
	if err != nil {
		return fmt.Errorf("input prompt failed: %w", err)
	}

	fmt.Println(color.CyanString("Executing prompt..."))

	// Execute prompt
	var result string
	if useClaudeCode {
		color.Cyan("Using Claude Code...")
		claudeClient := claude.NewClient()
		result, err = claudeClient.ExecutePrompt(string(promptContent), userInput)
		if err != nil {
			return fmt.Errorf("failed to execute prompt with Claude Code: %w", err)
		}
	} else {
		color.Cyan("Using OpenRouter...")
		client := openrouter.NewClient(apiKey)
		result, err = client.ExecutePrompt(string(promptContent), userInput)
		if err != nil {
			return fmt.Errorf("failed to execute prompt with OpenRouter: %w", err)
		}
	}

	color.Green("✓ Prompt executed successfully!\n")

	// Display response
	fmt.Println()
	color.Cyan("Response:")
	fmt.Println("─────────────────────────────────────────")
	fmt.Println(result)
	fmt.Println("─────────────────────────────────────────")

	// Ask if user wants to save the output
	promptSave := promptui.Prompt{
		Label:     "Would you like to save this output",
		IsConfirm: true,
		Default:   "y",
	}

	_, err = promptSave.Run()
	if err != nil {
		// User declined to save
		return nil
	}

	// Select output location
	locations := []string{
		"Project Overview (99_Assets/Project_Overview)",
		"Communications (99_Assets/Communications)",
		"POC Documents (99_Assets/POC_Documents)",
		"Notes (00_Inbox/notes)",
		"Other (specify)",
	}

	locationSelect := promptui.Select{
		Label: "Where would you like to save the output?",
		Items: locations,
	}

	locIdx, _, err := locationSelect.Run()
	if err != nil {
		return nil
	}

	var savePath string
	switch locIdx {
	case 0:
		savePath = "99_Assets/Project_Overview"
	case 1:
		savePath = "99_Assets/Communications"
	case 2:
		savePath = "99_Assets/POC_Documents"
	case 3:
		savePath = "00_Inbox/notes"
	case 4:
		promptCustom := promptui.Prompt{
			Label:   "Enter the path (relative to project root)",
			Default: "99_Assets",
		}
		customPath, err := promptCustom.Run()
		if err != nil {
			return nil
		}
		savePath = customPath
	}

	// Get filename
	defaultFilename := strings.TrimSuffix(selectedPrompt, ".md") + "_" + time.Now().Format("2006-01-02")
	promptFilename := promptui.Prompt{
		Label:   "Enter filename (without extension)",
		Default: defaultFilename,
		Validate: func(input string) error {
			if strings.TrimSpace(input) == "" {
				return fmt.Errorf("filename is required")
			}
			return nil
		},
	}

	filename, err := promptFilename.Run()
	if err != nil {
		return nil
	}

	// Create output content
	outputContent := fmt.Sprintf(`# %s

**Date:** %s
**Prompt Template:** %s
**Model:** %s

## User Input

%s

## Response

%s
`, strings.ReplaceAll(filename, "_", " "),
		time.Now().Format("2006-01-02 15:04:05"),
		selectedPrompt,
		"google/gemini-2.0-flash-exp:free",
		userInput,
		result)

	// Save to file
	fullPath := filepath.Join(".", savePath, filename+".md")
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(fullPath, []byte(outputContent), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	color.Green("✓ Output saved to: %s", fullPath)

	return nil
}
