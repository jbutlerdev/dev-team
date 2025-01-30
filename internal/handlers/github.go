package handlers

import (
	"dev-team/internal/state"
	"dev-team/pkg/github"
	"dev-team/pkg/repository"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"time"
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
		issues = append(issues, issue)
	}

	json.NewEncoder(w).Encode(issues)
}

func TrackIssues() {
    state.State.Mu.RLock()
    token := state.State.Settings.GitHubToken
    state.State.Mu.RUnlock()

    if token == "" {
        log.Println("GitHub token not configured, cannot track issues")
        return
    }

    for _, repo := range state.State.Repositories {
        issues, err := github.FetchIssues(token, repo.Owner, repo.Name, "dev-team")
        if err != nil {
            log.Printf("Error fetching issues for %s/%s: %v", repo.Owner, repo.Name, err)
            continue
        }

        state.State.Mu.Lock()
        for _, issue := range issues {
            if _, exists := state.State.TrackedIssues[issue.GetID()]; !exists {
                state.State.TrackedIssues[issue.GetID()] = issue
                log.Printf("Start tracking issue: %d - %s", issue.GetID(), issue.GetTitle())
            }
        }
        state.State.Mu.Unlock()
    }

    //Check for closed issues
    state.State.Mu.Lock()
    for id, issue := range state.State.TrackedIssues {
        if issue.IsClosed() {
            delete(state.State.TrackedIssues, id)
            log.Printf("Stop tracking issue: %d - %s", issue.GetID(), issue.GetTitle())
        }
    }
    state.State.Mu.Unlock()
}


func StartIssueTracking() {
    log.Println("Starting issue tracking")
    err := state.State.Scheduler.AddTask("issue-tracking", "*/5 * * * *", func() {
        TrackIssues()
    })
    if err != nil {
        log.Printf("Error adding issue tracking task: %v", err)
    }
}

func StopIssueTracking() {
    log.Println("Stopping issue tracking")
    state.State.Scheduler.RemoveTask("issue-tracking")
}
