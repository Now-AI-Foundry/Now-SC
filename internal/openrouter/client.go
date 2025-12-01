package openrouter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	OpenRouterAPIURL = "https://openrouter.ai/api/v1/chat/completions"
	DefaultModel     = "google/gemini-2.0-flash-exp:free"
)

type Client struct {
	apiKey string
	client *http.Client
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Request struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Response struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// NewClient creates a new OpenRouter client
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		client: &http.Client{},
	}
}

// ExecutePrompt executes a prompt using OpenRouter API
func (c *Client) ExecutePrompt(promptContent, userInput string) (string, error) {
	if userInput == "" {
		userInput = "Please provide guidance based on the system prompt."
	}

	reqBody := Request{
		Model: DefaultModel,
		Messages: []Message{
			{
				Role:    "system",
				Content: promptContent,
			},
			{
				Role:    "user",
				Content: userInput,
			},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", OpenRouterAPIURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("HTTP-Referer", "https://github.com/now-sc-cli")
	req.Header.Set("X-Title", "Now-SC CLI Tool")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("OpenRouter API error (status %d): %s", resp.StatusCode, string(body))
	}

	var apiResp Response
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(apiResp.Choices) == 0 {
		return "", fmt.Errorf("no response from API")
	}

	return apiResp.Choices[0].Message.Content, nil
}
