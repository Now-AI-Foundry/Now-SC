package claude

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// Client wraps Claude Code CLI interactions
type Client struct{}

// NewClient creates a new Claude Code client
func NewClient() *Client {
	return &Client{}
}

// ExecutePrompt sends a prompt to Claude Code via stdio and returns the response
func (c *Client) ExecutePrompt(promptContent, userInput string) (string, error) {
	// Combine prompt and user input
	fullPrompt := fmt.Sprintf("%s\n\nUser Request:\n%s", promptContent, userInput)

	// Execute claude code command
	cmd := exec.Command("claude", "code", "--stdio")

	// Setup stdin
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	// Setup stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Start the command
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start claude code: %w (is Claude Code installed?)", err)
	}

	// Write the prompt to stdin
	if _, err := io.WriteString(stdin, fullPrompt); err != nil {
		return "", fmt.Errorf("failed to write prompt: %w", err)
	}
	stdin.Close()

	// Wait for completion
	if err := cmd.Wait(); err != nil {
		return "", fmt.Errorf("claude code execution failed: %w\nStderr: %s", err, stderr.String())
	}

	response := stdout.String()
	if response == "" {
		return "", fmt.Errorf("no response from Claude Code")
	}

	return strings.TrimSpace(response), nil
}

// IsAvailable checks if Claude Code is installed and accessible
func IsAvailable() bool {
	cmd := exec.Command("claude", "code", "--version")
	err := cmd.Run()
	return err == nil
}

// ExecuteWithFiles executes a prompt with file contents as context
func (c *Client) ExecuteWithFiles(promptContent string, files []string, userInput string) (string, error) {
	var contextBuilder strings.Builder

	contextBuilder.WriteString("Context Files:\n\n")

	// Read and append file contents
	for _, filePath := range files {
		// Read file content (implementation needed)
		contextBuilder.WriteString(fmt.Sprintf("=== File: %s ===\n", filePath))
		// TODO: Read actual file content
		contextBuilder.WriteString("\n\n")
	}

	// Combine everything
	fullPrompt := fmt.Sprintf("%s\n\n%s\n\nUser Request:\n%s",
		contextBuilder.String(),
		promptContent,
		userInput)

	return c.ExecutePrompt(fullPrompt, "")
}

// StreamExecute executes a prompt and streams the response
func (c *Client) StreamExecute(promptContent, userInput string, writer io.Writer) error {
	fullPrompt := fmt.Sprintf("%s\n\nUser Request:\n%s", promptContent, userInput)

	cmd := exec.Command("claude", "code", "--stdio")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	cmd.Stdout = writer
	cmd.Stderr = writer

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start claude code: %w", err)
	}

	if _, err := io.WriteString(stdin, fullPrompt); err != nil {
		return fmt.Errorf("failed to write prompt: %w", err)
	}
	stdin.Close()

	return cmd.Wait()
}

// ReadStreamResponse reads a streamed response line by line
func ReadStreamResponse(reader io.Reader) (string, error) {
	var response strings.Builder
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		line := scanner.Text()
		response.WriteString(line)
		response.WriteString("\n")
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading stream: %w", err)
	}

	return strings.TrimSpace(response.String()), nil
}
