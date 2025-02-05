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

func HandleGitHubWebhook(w http.ResponseWriter, r *http.Request) {
	var payload map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, "Error decoding webhook payload", http.StatusBadRequest)
		return
	}

	eventType := r.Header.Get("X-GitHub-Event")

	switch eventType {
	case "pull_request":
		handlePullRequestEvent(payload)
	case "issue_comment":
		handleIssueCommentEvent(payload)
	default:
		fmt.Println("Unhandled event type:", eventType)
	}

	w.WriteHeader(http.StatusOK)
}

func handlePullRequestEvent(payload map[string]interface{}) {
	action, ok := payload["action"].(string)
	if !ok {
		fmt.Println("Error getting action from payload")
		return
	}

	pr, ok := payload["pull_request"].(map[string]interface{})
	if !ok {
		fmt.Println("Error getting pull_request from payload")
		return
	}

	number, ok := pr["number"].(float64)
	if !ok {
		fmt.Println("Error getting number from payload")
		return
	}
	numberInt := int(number)

	labelAdded := false
	labelRemoved := false

	if action == "labeled" {
		label, ok := payload["label"].(map[string]interface{})
		if !ok {
			fmt.Println("Error getting label from payload")
			return
		}
		labelName, ok := label["name"].(string)
		if !ok {
			fmt.Println("Error getting label name from payload")
			return
		}
		if labelName == "dev-team" {
			labelAdded = true
		}
	}

	if action == "unlabeled" {
		label, ok := payload["label"].(map[string]interface{})
		if !ok {
			fmt.Println("Error getting label from payload")
			return
		}
		labelName, ok := label["name"].(string)
		if !ok {
			fmt.Println("Error getting label name from payload")
			return
		}
		if labelName == "dev-team" {
			labelRemoved = true
		}
	}

	state.State.Mu.Lock()
	defer state.State.Mu.Unlock()

	_, isTracked := state.State.TrackedPRs[numberInt]

	switch action {
	case "opened", "synchronize", "reopened":
		if labelAdded || hasDevTeamLabel(pr) {
			state.State.TrackedPRs[numberInt] = pr
			fmt.Printf("PR #%d added to tracked PRs\n", numberInt)
		}
	case "closed":
		delete(state.State.TrackedPRs, numberInt)
		fmt.Printf("PR #%d removed from tracked PRs (closed)\n", numberInt)
	case "labeled":
		if labelAdded {
			state.State.TrackedPRs[numberInt] = pr
			fmt.Printf("PR #%d added to tracked PRs (labeled)\n", numberInt)
		}
	case "unlabeled":
		if labelRemoved {
			delete(state.State.TrackedPRs, numberInt)
			fmt.Printf("PR #%d removed from tracked PRs (unlabeled)\n", numberInt)
		}
	}

	// If PR is already tracked and dev-team label is removed then remove it from tracked PRs
	if isTracked && labelRemoved {
		delete(state.State.TrackedPRs, numberInt)
		fmt.Printf("PR #%d removed from tracked PRs (unlabeled)\n", numberInt)
	}
}

func handleIssueCommentEvent(payload map[string]interface{}) {
	issue, ok := payload["issue"].(map[string]interface{})
	if !ok {
		fmt.Println("Error getting issue from payload")
		return
	}

	number, ok := issue["number"].(float64)
	if !ok {
		fmt.Println("Error getting number from payload")
		return
	}
	numberInt := int(number)

	comment, ok := payload["comment"].(map[string]interface{})
	if !ok {
		fmt.Println("Error getting comment from payload")
		return
	}

	body, ok := comment["body"].(string)
	if !ok {
		fmt.Println("Error getting body from payload")
		return
	}

	state.State.Mu.RLock()
	_, isTracked := state.State.TrackedPRs[numberInt]
	state.State.Mu.RUnlock()

	if isTracked {
		// TODO: Add a commit to the PR addressing the comment
		fmt.Printf("Comment added to tracked PR #%d: %s\n", numberInt, body)
		pr, ok := payload["issue"].(map[string]interface{})
		if !ok {
			fmt.Println("Error getting pull_request from payload")
			return
		}

		repoURL, ok := pr["repository_url"].(string)
		if !ok {
			fmt.Println("Error getting repository_url from payload")
			return
		}

		repoName := filepath.Base(repoURL)

		// Get owner of the repo
		ownerURL, ok := pr["user"].(map[string]interface{})["url"].(string)
		if !ok {
			fmt.Println("Error getting owner url from payload")
			return
		}

		ownerName := filepath.Base(ownerURL)
		state.State.Mu.RLock()
		token := state.State.Settings.GitHubToken
		state.State.Mu.RUnlock()
		msg := fmt.Sprintf("Addressed comment: %s", body)
		err := github.AddCommitToPR(ownerName, repoName, numberInt, token, msg)
		if err != nil {
			fmt.Printf("Error adding commit to PR #%d: %v\n", numberInt, err)
			return
		}
	}
}

func hasDevTeamLabel(pr map[string]interface{}) bool {
	labels, ok := pr["labels"].([]interface{})
	if !ok {
		return false
	}

	for _, label := range labels {
		labelMap, ok := label.(map[string]interface{})
		if !ok {
			continue
		}
		name, ok := labelMap["name"].(string)
		if !ok {
			continue
		}
		if name == "dev-team" {
			return true
		}
	}

	return false
}
