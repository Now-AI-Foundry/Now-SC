package project

import (
	"fmt"
	"os"
	"path/filepath"
)

// DirectoryStructure defines the project directory layout
var DirectoryStructure = map[string]interface{}{
	"00_Inbox": map[string]interface{}{
		"calls": map[string]interface{}{
			"internal": map[string]interface{}{},
			"external": map[string]interface{}{},
		},
		"emails": map[string]interface{}{},
		"notes":  map[string]interface{}{},
	},
	"01_Customers":              map[string]interface{}{},
	"10_PromptTemplates":        map[string]interface{}{},
	"20_Demo_Library":           map[string]interface{}{},
	"30_CommunicationTemplates": map[string]interface{}{},
	"99_Assets": map[string]interface{}{
		"Project_Overview": map[string]interface{}{},
		"Communications":   map[string]interface{}{},
		"POC_Documents":    map[string]interface{}{},
	},
}

// CreateStructure creates the project directory structure
func CreateStructure(basePath, customerName string) error {
	return createDirectoryStructure(basePath, DirectoryStructure, customerName)
}

func createDirectoryStructure(basePath string, structure map[string]interface{}, customerName string) error {
	for dirName, subDirs := range structure {
		fullPath := filepath.Join(basePath, dirName)

		// Handle customer placeholder
		if dirName == "01_Customers" && customerName != "" {
			customerPath := filepath.Join(fullPath, customerName)
			if err := os.MkdirAll(customerPath, 0755); err != nil {
				return fmt.Errorf("failed to create customer directory: %w", err)
			}
		} else {
			if err := os.MkdirAll(fullPath, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", fullPath, err)
			}
		}

		// Recursively create subdirectories
		if subStructure, ok := subDirs.(map[string]interface{}); ok && len(subStructure) > 0 {
			if err := createDirectoryStructure(fullPath, subStructure, ""); err != nil {
				return err
			}
		}
	}
	return nil
}

// CreateProjectFiles creates the README, .env.example, and .gitignore files
func CreateProjectFiles(projectPath, projectName, customerName string) error {
	// Create README
	readmeContent := fmt.Sprintf(`# %s

## Customer: %s

This project was bootstrapped with Now-SC CLI tool.

## Directory Structure

- **00_Inbox/** - Raw meeting notes and transcripts
  - calls/internal - Internal call recordings and notes
  - calls/external - External call recordings and notes
  - emails - Email communications
  - notes - General notes

- **01_Customers/%s/** - Customer-specific information

- **10_PromptTemplates/** - Ready-to-use prompt templates

- **20_Demo_Library/** - Demo materials and resources

- **99_Assets/** - Processed and synthesized outputs
  - Project_Overview - High-level project summaries
  - Communications - Prepared communications
  - POC_Documents - Proof of concept documentation

## Using Prompts

To execute a prompt, use:
` + "```bash\nnow-sc prompt\n```" + `

Make sure you have set the OPENROUTER_API_KEY environment variable.
`, projectName, customerName, customerName)

	if err := os.WriteFile(filepath.Join(projectPath, "README.md"), []byte(readmeContent), 0644); err != nil {
		return fmt.Errorf("failed to create README: %w", err)
	}

	// Create .env.example
	envExample := `# OpenRouter API Key
# Get your API key from https://openrouter.ai/
OPENROUTER_API_KEY=your_api_key_here
`
	if err := os.WriteFile(filepath.Join(projectPath, ".env.example"), []byte(envExample), 0644); err != nil {
		return fmt.Errorf("failed to create .env.example: %w", err)
	}

	// Create .gitignore
	gitignoreContent := `node_modules/
.env
.DS_Store
*.log
`
	if err := os.WriteFile(filepath.Join(projectPath, ".gitignore"), []byte(gitignoreContent), 0644); err != nil {
		return fmt.Errorf("failed to create .gitignore: %w", err)
	}

	return nil
}
