package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/jbutlerdev/dev-team/internal/state"
	"github.com/jbutlerdev/dev-team/pkg/github"
	"github.com/jbutlerdev/dev-team/pkg/repository"
)

func HandleCreatePullRequest(w http.ResponseWriter, r *http.Request) {
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

	err = github.CreateDraftPR(repo.RemotePath, settings.GitHubToken, github.GitHubPRInput{
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
