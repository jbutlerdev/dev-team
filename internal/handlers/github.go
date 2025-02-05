package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	githubv53 "github.com/google/go-github/v53/github"
	"github.com/jbutlerdev/dev-team/internal/state"
	githubpkg "github.com/jbutlerdev/dev-team/pkg/github"
	"github.com/jbutlerdev/dev-team/pkg/repository"
)

func HandleGitHubRepositories(w http.ResponseWriter, r *http.Request) {
	state.State.Mu.RLock()
	token := state.State.Settings.GitHubToken
	state.State.Mu.RUnlock()

	repos, err := githubpkg.FetchRepositories(token)
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

func HandleGitHubWebhook(w http.ResponseWriter, r *http.Request) {
	eventType := r.Header.Get("X-GitHub-Event")
	payload, err := githubpkg.ValidateWebhook(r, state.State.Settings.GitHubWebhookSecret)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	switch eventType {
	case "pull_request":
		var pl githubv53.PullRequestEvent
		err = json.Unmarshal(payload, &pl)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		handlePullRequestEvent(pl)
	case "pull_request_review_comment":
		var prc githubv53.PullRequestReviewCommentEvent
		err = json.Unmarshal(payload, &prc)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		handlePullRequestReviewCommentEvent(prc)
	case "issues":
		var issueEvent githubv53.IssuesEvent
		err = json.Unmarshal(payload, &issueEvent)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		handleIssuesEvent(issueEvent)
	default:
		w.WriteHeader(http.StatusOK)
	}
}

func handlePullRequestEvent(pl githubv53.PullRequestEvent) {
	if pl.PullRequest == nil || pl.PullRequest.HTMLURL == nil {
		return
	}
	pullRequestURL := pl.PullRequest.GetHTMLURL()

	isDevTeamLabel := false
	for _, label := range pl.PullRequest.Labels {
		if label.GetName() == "dev-team" {
			isDevTeamLabel = true
			break
		}
	}

	switch pl.GetAction() {
	case "opened", "reopened", "labeled", "synchronize":
		if isDevTeamLabel {
			state.State.TrackPullRequest(pullRequestURL)
		}
	case "closed":
		state.State.UntrackPullRequest(pullRequestURL)
	}
}

func handlePullRequestReviewCommentEvent(prc githubv53.PullRequestReviewCommentEvent) {
	if prc.PullRequest == nil || prc.PullRequest.HTMLURL == nil {
		return
	}
	pullRequestURL := prc.PullRequest.GetHTMLURL()

	if state.State.IsTracked(pullRequestURL) {
		// Extract owner and repo name from the pull request URL
		parts := strings.Split(prc.PullRequest.GetHTMLURL(), "/")
		owner := parts[3]
		repo := parts[4]
		pullRequestNumber := prc.PullRequest.GetNumber()

		// Construct the commit message and content
		commitMessage := fmt.Sprintf("Addressed comment: %s", prc.Comment.GetBody())
		content := fmt.Sprintf("Addressed comment: %s", prc.Comment.GetBody())
		// file path
		path := prc.Comment.GetPath()

		//Get Token
		state.State.Mu.RLock()
		token := state.State.Settings.GitHubToken
		state.State.Mu.RUnlock()

		// Get the branch name
		branch := prc.PullRequest.GetHead().GetRef()

		// Add the commit to the pull request
		err := githubpkg.AddCommitToPullRequest(token, owner, repo, pullRequestNumber, commitMessage, content, path, branch)
		if err != nil {
			fmt.Println("Error adding commit to pull request:", err)
		}
		fmt.Println("Adding commit to PR", pullRequestURL)
	}
}

func handleIssuesEvent(issueEvent githubv53.IssuesEvent) {
	if issueEvent.Issue == nil || issueEvent.Issue.HTMLURL == nil {
		return
	}
	issueURL := issueEvent.Issue.GetHTMLURL()

	isDevTeamLabel := false
	for _, label := range issueEvent.Issue.Labels {
		if label.GetName() == "dev-team" {
			isDevTeamLabel = true
			break
		}
	}

	switch issueEvent.GetAction() {
	case "labeled":
		if isDevTeamLabel {
			state.State.TrackPullRequest(issueURL)
		}
	case "unlabeled", "closed":
		state.State.UntrackPullRequest(issueURL)
	}
}
