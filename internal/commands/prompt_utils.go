package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PromptInfo contains information about a prompt template
type PromptInfo struct {
	Name        string // Display name (without .md extension)
	FileName    string // Actual filename
	Path        string // Full path to file
	Description string // First line of the prompt (if available)
}

// ListPrompts returns all available prompt templates
func ListPrompts(projectRoot string) ([]PromptInfo, error) {
	promptsPath := filepath.Join(projectRoot, "10_PromptTemplates")

	if _, err := os.Stat(promptsPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("prompt templates directory not found at: %s", promptsPath)
	}

	entries, err := os.ReadDir(promptsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read prompts directory: %w", err)
	}

	var prompts []PromptInfo
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		fullPath := filepath.Join(promptsPath, entry.Name())
		name := strings.TrimSuffix(entry.Name(), ".md")
		name = strings.ReplaceAll(name, "_", " ")

		// Try to read first line for description
		description := ""
		if content, err := os.ReadFile(fullPath); err == nil {
			lines := strings.Split(string(content), "\n")
			if len(lines) > 0 {
				description = strings.TrimSpace(strings.TrimPrefix(lines[0], "#"))
				if len(description) > 60 {
					description = description[:60] + "..."
				}
			}
		}

		prompts = append(prompts, PromptInfo{
			Name:        name,
			FileName:    entry.Name(),
			Path:        fullPath,
			Description: description,
		})
	}

	if len(prompts) == 0 {
		return nil, fmt.Errorf("no prompt templates found in: %s", promptsPath)
	}

	return prompts, nil
}

// FindPrompt finds a prompt by name (fuzzy matching)
func FindPrompt(projectRoot, promptName string) (*PromptInfo, error) {
	prompts, err := ListPrompts(projectRoot)
	if err != nil {
		return nil, err
	}

	// Normalize search term
	searchTerm := strings.ToLower(strings.TrimSpace(promptName))

	// Exact match first
	for _, prompt := range prompts {
		if strings.ToLower(prompt.Name) == searchTerm {
			return &prompt, nil
		}
		if strings.ToLower(strings.TrimSuffix(prompt.FileName, ".md")) == searchTerm {
			return &prompt, nil
		}
	}

	// Partial match
	for _, prompt := range prompts {
		if strings.Contains(strings.ToLower(prompt.Name), searchTerm) {
			return &prompt, nil
		}
	}

	return nil, fmt.Errorf("prompt not found: %s", promptName)
}

// FileInfo contains information about a discovered file
type FileInfo struct {
	Path         string // Full path
	RelativePath string // Path relative to project root
	Name         string // Filename
	Size         int64  // File size in bytes
	ModTime      string // Last modified time
}

// DiscoverFiles scans the inbox folders for files
func DiscoverFiles(projectRoot string) ([]FileInfo, error) {
	inboxPath := filepath.Join(projectRoot, "00_Inbox")

	if _, err := os.Stat(inboxPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("inbox directory not found at: %s", inboxPath)
	}

	var files []FileInfo

	err := filepath.Walk(inboxPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(projectRoot, path)
		if err != nil {
			relPath = path
		}

		files = append(files, FileInfo{
			Path:         path,
			RelativePath: relPath,
			Name:         info.Name(),
			Size:         info.Size(),
			ModTime:      info.ModTime().Format("2006-01-02 15:04"),
		})

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan inbox: %w", err)
	}

	return files, nil
}

// ReadFileContent reads and returns the content of a file as a string
func ReadFileContent(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	return string(content), nil
}

// FormatFileContext formats file contents for inclusion in a prompt
func FormatFileContext(files []string) (string, error) {
	var builder strings.Builder

	builder.WriteString("Context Files:\n\n")

	for _, filePath := range files {
		content, err := ReadFileContent(filePath)
		if err != nil {
			return "", err
		}

		builder.WriteString(fmt.Sprintf("=== File: %s ===\n\n", filepath.Base(filePath)))
		builder.WriteString(content)
		builder.WriteString("\n\n")
	}

	return builder.String(), nil
}
