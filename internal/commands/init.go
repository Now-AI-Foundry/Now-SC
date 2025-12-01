package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Now-AI-Foundry/Now-SC/internal/github"
	"github.com/Now-AI-Foundry/Now-SC/internal/project"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var (
	projectName  string
	customerName string
	noGitHub     bool
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new presales project",
	Long: `Creates a new presales project with the standard directory structure,
fetches base prompts from GitHub, and optionally creates a GitHub repository.`,
	RunE: runInit,
}

func init() {
	initCmd.Flags().StringVarP(&projectName, "name", "n", "", "Project name")
	initCmd.Flags().StringVarP(&customerName, "customer", "c", "", "Customer name")
	initCmd.Flags().BoolVar(&noGitHub, "no-github", false, "Skip GitHub repository creation")
}

func runInit(cmd *cobra.Command, args []string) error {
	// Interactive prompts if flags not provided
	if projectName == "" {
		prompt := promptui.Prompt{
			Label:   "Project name",
			Default: "presales-project",
		}
		result, err := prompt.Run()
		if err != nil {
			return fmt.Errorf("prompt failed: %w", err)
		}
		projectName = result
	}

	if customerName == "" {
		prompt := promptui.Prompt{
			Label: "Customer name",
			Validate: func(input string) error {
				if input == "" {
					return fmt.Errorf("customer name is required")
				}
				return nil
			},
		}
		result, err := prompt.Run()
		if err != nil {
			return fmt.Errorf("prompt failed: %w", err)
		}
		customerName = result
	}

	projectPath := filepath.Join(".", projectName)

	// Check if directory already exists
	if _, err := os.Stat(projectPath); err == nil {
		prompt := promptui.Prompt{
			Label:     fmt.Sprintf("Directory %s already exists. Overwrite", projectName),
			IsConfirm: true,
		}
		_, err := prompt.Run()
		if err != nil {
			color.Yellow("Project initialization cancelled.")
			return nil
		}

		if err := os.RemoveAll(projectPath); err != nil {
			return fmt.Errorf("failed to remove existing directory: %w", err)
		}
	}

	// Create project structure
	fmt.Println(color.CyanString("Creating project structure..."))
	if err := project.CreateStructure(projectPath, customerName); err != nil {
		return fmt.Errorf("failed to create project structure: %w", err)
	}

	// Fetch prompts from GitHub
	fmt.Println(color.CyanString("Fetching base prompts from GitHub..."))
	if err := github.FetchAndSavePrompts(projectPath); err != nil {
		return fmt.Errorf("failed to fetch prompts: %w", err)
	}

	// Fetch communication templates
	fmt.Println(color.CyanString("Fetching communication templates..."))
	if err := github.FetchCommunicationTemplates(projectPath); err != nil {
		color.Yellow("\nWarning: Failed to fetch some templates")
	}

	// Create project files
	if err := project.CreateProjectFiles(projectPath, projectName, customerName); err != nil {
		return fmt.Errorf("failed to create project files: %w", err)
	}

	color.Green("✓ Project \"%s\" created successfully!\n", projectName)

	// Create GitHub repository if not skipped
	githubToken := os.Getenv("GITHUB_PAT")
	if !noGitHub && githubToken != "" {
		fmt.Println(color.CyanString("Creating GitHub repository..."))
		if err := github.CreateRepository(projectPath, projectName, customerName); err != nil {
			color.Red("✗ GitHub repository creation failed: %v", err)
			color.Yellow("You can create the repository manually later.")
		} else {
			color.Green("✓ GitHub repository created!")
		}
	} else if noGitHub {
		fmt.Println("\nSkipped GitHub repository creation.")
	} else {
		color.Yellow("\nNote: GITHUB_PAT environment variable not set. Skipping GitHub repository creation.")
		fmt.Println("To enable automatic repository creation, set your GitHub Personal Access Token:")
		fmt.Println("  export GITHUB_PAT=your_token_here")
	}

	// Print summary
	fmt.Println()
	color.Cyan("Project structure created:")
	fmt.Printf("  %s/\n", projectPath)
	fmt.Println("  ├── 00_Inbox/")
	fmt.Printf("  ├── 01_Customers/%s/\n", customerName)
	fmt.Println("  ├── 10_PromptTemplates/")
	fmt.Println("  ├── 20_Demo_Library/")
	fmt.Println("  └── 99_Assets/")

	fmt.Println()
	color.Yellow("Next steps:")
	fmt.Printf("  1. cd %s\n", projectName)
	fmt.Println("  2. Set your OPENROUTER_API_KEY environment variable")
	fmt.Println("  3. Run \"now-sc prompt\" to execute prompts")

	return nil
}
