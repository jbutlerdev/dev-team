package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/jbutlerdev/dev-team/internal/state"
	"github.com/jbutlerdev/dev-team/pkg/repository"
)

type Repository struct {
	Name string `json:"name"`
}

type Issue struct {
	Number    int    `json:"number"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	HTMLURL   string `json:"html_url"`
	SourceURL string `json:"url"`
	State     string `json:"state"`
}

type PullRequest struct {
	Number          int      `json:"number"`
	Title           string   `json:"title"`
	Body            string   `json:"body"`
	CreatedAt       string   `json:"created_at"`
	UpdatedAt       string   `json:"updated_at"`
	Labels          []string `json:"labels"`
	IssueURL        string   `json:"issue_url"`
	LinkedIssueURLs []string `json:"linked_issue_urls"`
}

type Comment struct {
	ID        int    `json:"id"`
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func FetchRepositories(token string) ([]Repository, error) {
	url := "https://api.github.com/user/repos"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var repos []Repository
	err = json.Unmarshal(body, &repos)
	if err != nil {
		return nil, err
	}

	return repos, nil
}

func FetchIssues(repoPath, org, token string) ([]Issue, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues", org, repoPath)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var issues []Issue
	err = json.Unmarshal(body, &issues)
	if err != nil {
		return nil, err
	}

	return issues, nil
}

func FetchPullRequests(repoPath, org, token string) ([]PullRequest, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls", org, repoPath)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var pullRequests []PullRequest
	err = json.Unmarshal(body, &pullRequests)
	if err != nil {
		return nil, err
	}

	return pullRequests, nil
}

func CreateDraftPR(repoPath, token string, input GitHubPRInput) error {
	url := fmt.Sprintf("https://api.github.com/repos/dev-team/%s/pulls", repoPath)

	payload, err := json.Marshal(input)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(string(payload)))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 300 {
		return fmt.Errorf("failed to create pull request: %s", string(body))
	}

	return nil
}

type GitHubPRInput struct {
	Title               string `json:"title"`
	Head                string `json:"head"`
	Base                string `json:"base"`
	Body                string `json:"body"`
	Draft               bool   `json:"draft"`
	MaintainerCanModify bool   `json:"maintainer_can_modify"`
}

func FetchPullRequestComments(repoPath string, pullRequestNumber int, token string) ([]Comment, error) {
	url := fmt.Sprintf("https://api.github.com/repos/dev-team/%s/issues/%d/comments", repoPath, pullRequestNumber)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var comments []Comment
	err = json.Unmarshal(body, &comments)
	if err != nil {
		return nil, err
	}

	return comments, nil
}

func HandleGitHubRepositories(w http.ResponseWriter, r *http.Request) {
	state.State.Mu.RLock()
	token := state.State.Settings.GitHubToken
	state.State.Mu.RUnlock()

	repos, err := FetchRepositories(token)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching repositories: %v", err), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(repos)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func HandleGitHubIssues(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "Path parameter is required", http.StatusBadRequest)
		return
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	state.State.Mu.RLock()
	repo, exists := state.State.Repositories[absPath]
	state.State.Mu.RUnlock()

	if !exists {
		http.Error(w, "Repository not found", http.StatusNotFound)
		return
	}

	state.State.Mu.RLock()
	token := state.State.Settings.GitHubToken
	state.State.Mu.RUnlock()

	if token == "" {
		http.Error(w, "GitHub token not configured", http.StatusBadRequest)
		return
	}

	err = repo.UpdateIssues(token)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching issues: %v", err), http.StatusInternalServerError)
		return
	}

	issues := make([]repository.Issue, 0, len(repo.Issues))
	for _, issue := range repo.Issues {
		issues = append(issues, *issue)
	}

	err = json.NewEncoder(w).Encode(issues)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func HandleCreatePR(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "Path parameter is required", http.StatusBadRequest)
		return
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	state.State.Mu.RLock()
	repo, exists := state.State.Repositories[absPath]
	settings := state.State.Settings
	state.State.Mu.RUnlock()

	if !exists {
		http.Error(w, "Repository not found", http.StatusNotFound)
		return
	}

	if settings.GitHubToken == "" {
		http.Error(w, "GitHub token not configured", http.StatusBadRequest)
		return
	}

	summary, err := repo.ChangeSummary()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting change summary: %v", err), http.StatusInternalServerError)
		return
	}

	commitMessage, err := state.State.GenAI.Generate(repository.MODEL, repository.CommitPrompt(summary))
	if err != nil {
		http.Error(w, fmt.Sprintf("Error generating commit message: %v", err), http.StatusInternalServerError)
		return
	}
	// Commit changes
	err = repo.Commit(commitMessage)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error committing changes: %v", err), http.StatusInternalServerError)
		return
	}

	// Push changes
	err = repo.Push()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error pushing changes: %v", err), http.StatusInternalServerError)
		return
	}

	prTitle, err := state.State.GenAI.Generate(repository.MODEL, repository.CommitPrompt(summary))
	if err != nil {
		http.Error(w, fmt.Sprintf("Error generating PR title: %v", err), http.StatusInternalServerError)
		return
	}

	prDescription, err := state.State.GenAI.Generate(repository.MODEL, repository.PRPrompt(summary))
	if err != nil {
		http.Error(w, fmt.Sprintf("Error generating PR description: %v", err), http.StatusInternalServerError)
		return
	}

	// add issue close tag to description
	//prDescription = fmt.Sprintf("%s\n\n%s\n<!--%s-->",
	//	prDescription,
	//	fmt.Sprintf("Closes #%d", issue.ID),
	//	issue.SourceURL)

	err = CreateDraftPR(repo.RemotePath, settings.GitHubToken, GitHubPRInput{
		Title:               prTitle,
		Head:                repo.State.CurrentBranch,
		Base:                "main",
		Body:                prDescription,
		Draft:               true,
		MaintainerCanModify: true,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating PR: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

}
