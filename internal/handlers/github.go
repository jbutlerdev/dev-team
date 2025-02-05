package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/jbutlerdev/dev-team/internal/state"
	"github.com/jbutlerdev/dev-team/pkg/github"
	"github.com/jbutlerdev/dev-team/pkg/repository"
)

func HandleGitHubRepositories(w http.ResponseWriter, r *http.Request) {
	state.State.Mu.RLock()
	token := state.State.Settings.GitHubToken
	state.State.Mu.RUnlock()

	repos, err := github.FetchRepositories(token)
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

func HandleGitHubPullRequestEvent(w http.ResponseWriter, r *http.Request) {
	var payload map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error decoding payload: %v", err), http.StatusBadRequest)
		return
	}

	action, ok := payload["action"].(string)
	if !ok {
		http.Error(w, "Missing action", http.StatusBadRequest)
		return
	}

	pr, ok := payload["pull_request"].(map[string]interface{})
	if !ok {
		http.Error(w, "Missing pull_request", http.StatusBadRequest)
		return
	}

	prIDFloat, ok := pr["id"].(float64)
	if !ok {
		http.Error(w, "Missing pull_request ID", http.StatusBadRequest)
		return
	}
	prID := int64(prIDFloat)

	state.State.Mu.Lock()
	defer state.State.Mu.Unlock()

	if state.State.TrackedPullRequests == nil {
		state.State.TrackedPullRequests = make(map[int64]bool)
	}

	switch action {
	case "labeled":
		label, ok := payload["label"].(map[string]interface{})
		if !ok {
			http.Error(w, "Missing label", http.StatusBadRequest)
			return
		}
		labelName, ok := label["name"].(string)
		if !ok {
			http.Error(w, "Missing label name", http.StatusBadRequest)
			return
		}
		if labelName == "dev-team" {
			state.State.TrackedPullRequests[prID] = true
		}
	case "unlabeled":
		label, ok := payload["label"].(map[string]interface{})
		if !ok {
			http.Error(w, "Missing label", http.StatusBadRequest)
			return
		}
		labelName, ok := label["name"].(string)
		if !ok {
			http.Error(w, "Missing label name", http.StatusBadRequest)
			return
		}
		if labelName == "dev-team" {
			delete(state.State.TrackedPullRequests, prID)
		}
	case "closed", "merged":
		delete(state.State.TrackedPullRequests, prID)
	case "opened", "synchronize":
		// do nothing
	default:
		fmt.Printf("Unhandled pull request event action: %s\n", action)
	}

	w.WriteHeader(http.StatusOK)
}
