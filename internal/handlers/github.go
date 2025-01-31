package handlers

import (
	"dev-team/internal/state"
	"dev-team/pkg/github"
	"dev-team/pkg/repository"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
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

	json.NewEncoder(w).Encode(repos)
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

	json.NewEncoder(w).Encode(issues)
}

func CreatePRFromIssueHandler(w http.ResponseWriter, r *http.Request) {
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

    issueTitle := r.URL.Query().Get("title")
    if issueTitle == "" {
        http.Error(w, "Issue title is required", http.StatusBadRequest)
        return
    }

    issueBody := r.URL.Query().Get("body")
    if issueBody == "" {
       http.Error(w, "Issue body is required", http.StatusBadRequest)
       return
    }

    baseBranch := r.URL.Query().Get("base")
    if baseBranch == "" {
        http.Error(w, "Base branch is required", http.StatusBadRequest)
        return
    }

    state.State.Mu.RLock()
    token := state.State.Settings.GitHubToken
    state.State.Mu.RUnlock()

    if token == "" {
        http.Error(w, "GitHub token not configured", http.StatusBadRequest)
        return
    }

    err = github.CreatePRFromIssue(absPath, issueTitle, issueBody, baseBranch, token)
    if err != nil {
        http.Error(w, fmt.Sprintf("Error creating PR from issue: %v", err), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    w.Write([]byte("PR created successfully"))
}
