package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	GitHubBaseURL       = "https://api.github.com/repos/Now-AI-Foundry/Now-SC-Base-Prompts/contents/Prompts"
	GitHubAPIURL        = "https://api.github.com"
	GitHubOrg           = "Now-AI-Foundry"
	TemplateBaseURL     = "https://raw.githubusercontent.com/Now-AI-Foundry/Now-SC-Base-Prompts/main/Templates"
)

type GitHubFile struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	DownloadURL string `json:"download_url"`
}

type GitHubRepo struct {
	CloneURL string `json:"clone_url"`
	HTMLURL  string `json:"html_url"`
}

// FetchAndSavePrompts fetches prompts from GitHub and saves them
func FetchAndSavePrompts(projectPath string) error {
	resp, err := http.Get(GitHubBaseURL)
	if err != nil {
		return fmt.Errorf("failed to fetch prompts: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var files []GitHubFile
	if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	promptsPath := filepath.Join(projectPath, "10_PromptTemplates")

	for _, file := range files {
		if file.Type == "file" && strings.HasSuffix(file.Name, ".md") {
			content, err := downloadFile(file.DownloadURL)
			if err != nil {
				return fmt.Errorf("failed to download %s: %w", file.Name, err)
			}

			filePath := filepath.Join(promptsPath, file.Name)
			if err := os.WriteFile(filePath, content, 0644); err != nil {
				return fmt.Errorf("failed to save %s: %w", file.Name, err)
			}
		}
	}

	return nil
}

// FetchCommunicationTemplates fetches communication templates
func FetchCommunicationTemplates(projectPath string) error {
	templates := []struct {
		URL      string
		Filename string
	}{
		{
			URL:      TemplateBaseURL + "/servicenow_poc_status_template.html",
			Filename: "servicenow_poc_status_template.html",
		},
	}

	templatesPath := filepath.Join(projectPath, "30_CommunicationTemplates")

	for _, template := range templates {
		content, err := downloadFile(template.URL)
		if err != nil {
			// Just warn, don't fail
			continue
		}

		filePath := filepath.Join(templatesPath, template.Filename)
		if err := os.WriteFile(filePath, content, 0644); err != nil {
			continue
		}
	}

	return nil
}

// CreateRepository creates a GitHub repository
func CreateRepository(projectPath, projectName, customerName string) error {
	token := os.Getenv("GITHUB_PAT")
	if token == "" {
		return fmt.Errorf("GITHUB_PAT environment variable not set")
	}

	// Sanitize repo name
	repoName := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '-'
	}, projectName)

	description := fmt.Sprintf("Presales project for %s", customerName)

	// Try to create in organization first
	repo, err := createOrgRepository(token, repoName, description)
	if err != nil {
		// If org creation fails due to permissions, try user account
		if strings.Contains(err.Error(), "403") || strings.Contains(err.Error(), "admin access") {
			fmt.Printf("Note: Cannot create in %s organization. Creating in your personal account instead...\n", GitHubOrg)
			repo, err = createUserRepository(token, repoName, description)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// Initialize git repository
	if err := initGitRepo(projectPath, repo.CloneURL); err != nil {
		return fmt.Errorf("failed to initialize git: %w", err)
	}

	fmt.Printf("Repository URL: %s\n", repo.HTMLURL)

	return nil
}

func createOrgRepository(token, repoName, description string) (*GitHubRepo, error) {
	reqBody := fmt.Sprintf(`{
		"name": "%s",
		"description": "%s",
		"private": true,
		"auto_init": false
	}`, repoName, description)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/orgs/%s/repos", GitHubAPIURL, GitHubOrg), strings.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnprocessableEntity {
		return nil, fmt.Errorf("repository \"%s\" already exists in %s organization", repoName, GitHubOrg)
	}

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, string(body))
	}

	var repo GitHubRepo
	if err := json.NewDecoder(resp.Body).Decode(&repo); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &repo, nil
}

func createUserRepository(token, repoName, description string) (*GitHubRepo, error) {
	reqBody := fmt.Sprintf(`{
		"name": "%s",
		"description": "%s",
		"private": true,
		"auto_init": false
	}`, repoName, description)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/user/repos", GitHubAPIURL), strings.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnprocessableEntity {
		body, _ := io.ReadAll(resp.Body)
		if strings.Contains(string(body), "already exists") {
			return nil, fmt.Errorf("repository \"%s\" already exists in your account", repoName)
		}
		return nil, fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, string(body))
	}

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, string(body))
	}

	var repo GitHubRepo
	if err := json.NewDecoder(resp.Body).Decode(&repo); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &repo, nil
}

func initGitRepo(projectPath, repoURL string) error {
	commands := [][]string{
		{"git", "init"},
		{"git", "remote", "add", "origin", repoURL},
		{"git", "branch", "-M", "main"},
	}

	for _, cmdArgs := range commands {
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		cmd.Dir = projectPath
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to run %s: %w", strings.Join(cmdArgs, " "), err)
		}
	}

	return nil
}

func downloadFile(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
