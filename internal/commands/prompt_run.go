package commands

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Now-AI-Foundry/Now-SC/internal/claude"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var (
	inputFiles      []string
	useClaudeCode   bool
	discoverFiles   bool
	saveOutput      bool
	outputPath      string
	stdinInput      bool
)

var promptRunCmd = &cobra.Command{
	Use:   "run <prompt-name>",
	Short: "Execute a specific prompt template",
	Long: `Execute a specific prompt template by name. Can accept input via stdin (pipes),
files, or interactive prompt.

Examples:
  # With piped input
  cat discovery.txt | now-sc prompt run sales-discovery

  # With file input
  now-sc prompt run sales-discovery --file inbox/notes/meeting.txt

  # Interactive input
  now-sc prompt run sales-discovery

  # Auto-discover files from inbox
  now-sc prompt run sales-discovery --discover
`,
	Args: cobra.ExactArgs(1),
	RunE: runPromptRun,
}

func init() {
	promptRunCmd.Flags().StringSliceVarP(&inputFiles, "file", "f", []string{}, "Input file(s) to include as context")
	promptRunCmd.Flags().BoolVar(&useClaudeCode, "claude", true, "Use Claude Code instead of OpenRouter (default: true)")
	promptRunCmd.Flags().BoolVar(&discoverFiles, "discover", false, "Auto-discover and select files from inbox")
	promptRunCmd.Flags().BoolVar(&saveOutput, "save", true, "Prompt to save output (default: true)")
	promptRunCmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output file path (skips save prompt)")
}

func runPromptRun(cmd *cobra.Command, args []string) error {
	promptName := args[0]
	projectRoot := "."

	// Find the prompt
	prompt, err := FindPrompt(projectRoot, promptName)
	if err != nil {
		color.Red("Error: %v", err)
		fmt.Println()
		color.Yellow("Available prompts:")
		if prompts, listErr := ListPrompts(projectRoot); listErr == nil {
			for _, p := range prompts {
				fmt.Printf("  - %s\n", p.Name)
			}
		}
		return err
	}

	// Read prompt content
	promptContent, err := os.ReadFile(prompt.Path)
	if err != nil {
		return fmt.Errorf("failed to read prompt file: %w", err)
	}

	color.Cyan("Using prompt: %s", prompt.Name)
	fmt.Println()

	// Get user input
	var userInput string
	var fileContext string

	// Check if stdin has data (piped input)
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		// Reading from pipe
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read stdin: %w", err)
		}
		userInput = string(data)
		color.Green("✓ Read input from stdin")
		fmt.Println()
	}

	// Handle file discovery
	if discoverFiles {
		selectedFiles, err := discoverAndSelectFiles(projectRoot)
		if err != nil {
			color.Yellow("Warning: %v", err)
		} else if len(selectedFiles) > 0 {
			inputFiles = append(inputFiles, selectedFiles...)
		}
	}

	// Read file contexts
	if len(inputFiles) > 0 {
		fileContext, err = FormatFileContext(inputFiles)
		if err != nil {
			return fmt.Errorf("failed to read context files: %w", err)
		}
		color.Green("✓ Loaded %d context file(s)", len(inputFiles))
		fmt.Println()
	}

	// If no input yet, prompt for it
	if userInput == "" && fileContext == "" {
		promptInput := promptui.Prompt{
			Label: "Enter your input for this prompt",
		}
		userInput, err = promptInput.Run()
		if err != nil {
			return fmt.Errorf("input prompt failed: %w", err)
		}
	}

	// Combine file context and user input
	fullInput := fileContext
	if userInput != "" {
		if fullInput != "" {
			fullInput += "\n\nUser Input:\n"
		}
		fullInput += userInput
	}

	// Execute prompt
	color.Cyan("Executing prompt...")
	fmt.Println()

	var result string
	if useClaudeCode {
		// Check if Claude Code is available
		if !claude.IsAvailable() {
			color.Red("Error: Claude Code is not installed or not in PATH")
			color.Yellow("Please install Claude Code or use --claude=false to use OpenRouter")
			return fmt.Errorf("Claude Code not available")
		}

		client := claude.NewClient()
		result, err = client.ExecutePrompt(string(promptContent), fullInput)
		if err != nil {
			return fmt.Errorf("failed to execute prompt with Claude Code: %w", err)
		}
	} else {
		// Fallback to OpenRouter (existing implementation)
		return fmt.Errorf("OpenRouter integration not yet implemented for run command")
	}

	// Display response
	fmt.Println()
	color.Cyan("Response:")
	fmt.Println("─────────────────────────────────────────")
	fmt.Println(result)
	fmt.Println("─────────────────────────────────────────")
	fmt.Println()

	color.Green("✓ Prompt executed successfully!")
	fmt.Println()

	// Handle output saving
	if outputPath != "" {
		// Direct output to specified path
		return savePromptOutput(projectRoot, prompt.Name, fullInput, result, outputPath)
	}

	if saveOutput {
		// Ask if user wants to save
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

		return savePromptOutputInteractive(projectRoot, prompt.Name, fullInput, result)
	}

	return nil
}

func discoverAndSelectFiles(projectRoot string) ([]string, error) {
	files, err := DiscoverFiles(projectRoot)
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no files found in inbox")
	}

	color.Cyan("Discovered %d file(s) in inbox:", len(files))
	fmt.Println()

	// Create selection items
	items := make([]string, len(files))
	for i, file := range files {
		sizeKB := file.Size / 1024
		items[i] = fmt.Sprintf("%s (%d KB, modified: %s)", file.RelativePath, sizeKB, file.ModTime)
	}

	// Select prompt (single select for now)
	// Note: promptui doesn't have native multi-select
	// TODO: Implement proper multi-select or use a different library
	prompt := promptui.Select{
		Label: "Select file to include as context",
		Items: items,
	}

	idx, _, err := prompt.Run()
	if err != nil {
		return nil, err
	}

	return []string{files[idx].Path}, nil
}

func savePromptOutput(projectRoot, promptName, input, response, outputPath string) error {
	// Create output content
	outputContent := fmt.Sprintf(`# %s

**Date:** %s
**Prompt:** %s

## Input

%s

## Response

%s
`, promptName,
		time.Now().Format("2006-01-02 15:04:05"),
		promptName,
		input,
		response)

	// Ensure directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(outputPath, []byte(outputContent), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	color.Green("✓ Output saved to: %s", outputPath)
	return nil
}

func savePromptOutputInteractive(projectRoot, promptName, input, response string) error {
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
	defaultFilename := strings.ReplaceAll(promptName, " ", "_") + "_" + time.Now().Format("2006-01-02")
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

	fullPath := filepath.Join(projectRoot, savePath, filename+".md")
	return savePromptOutput(projectRoot, promptName, input, response, fullPath)
}
