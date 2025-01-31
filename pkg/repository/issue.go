package repository

import (
	"dev-team/pkg/github"
	"fmt"
	"log"
	"strings"
	"regexp"
)

type Issue struct {
	ID          int          `json:"id"`
	Title       string       `json:"title"`
	Body        string       `json:"body"`
	Labels      []string     `json:"labels"`
	CreatedAt   string       `json:"created_at"`
	UpdatedAt   string       `json:"updated_at"`
	SourceURL   string       `json:"source_url"`
	PullRequest *PullRequest `json:"pull_request"`
	State       string       `json:"state"`
}

func (r *Repository) GetIssues() ([]Issue, error) {
	issues := make([]Issue, 0, len(r.Issues))
	for _, issue := range r.Issues {
		issues = append(issues, issue)
	}
	return issues, nil
}

func (r *Repository) UpdateIssues(token string) error {
	if r.RemotePath == "" {
		return fmt.Errorf("repository remote path is not set")
	}
	issues, err := github.FetchIssues(r.RemotePath, "dev-team", token)
	if err != nil {
		log.Printf("Error fetching issues: %v, request: %v", err, r.RemotePath)
		return err
	}
	for _, issue := range issues {
		r.Issues[issue.Number] = ghIssueToIssue(issue)
	}
	return nil
}

func (i *Issue) ToString() string {
	return fmt.Sprintf("Issue: %d\n\n"+
		"Title: %s\n\n"+
		"Body: %s\n\n", i.ID, i.Title, i.Body)
}

func ghIssueToIssue(issue github.Issue) Issue {
	return Issue{
		ID:        issue.Number,
		Title:     issue.Title,
		Body:      issue.Body,
		CreatedAt: issue.CreatedAt,
		UpdatedAt: issue.UpdatedAt,
		SourceURL: issue.HTMLURL,
		State:     issue.State,
	}
}

func (r *Repository) StartIssue(issue Issue) error {
    branchName := issue.Title
	// Convert to lowercase
	branchName = strings.ToLower(branchName)
	// Replace spaces with dashes
	re := regexp.MustCompile(`\s+`)
	branchName = re.ReplaceAllString(branchName, "-")
    // Truncate to 100 characters
    if len(branchName) > 100 {
        branchName = branchName[:100]
    }

    // Ensure main is up to date
    err := r.git.Checkout("main")
    if err != nil {
        return fmt.Errorf("failed to checkout main: %w", err)
    }
    err = r.git.Pull()
    if err != nil {
        return fmt.Errorf("failed to pull main: %w", err)
    }

    // Create the new branch
    err = r.git.CreateBranch(branchName)
    if err != nil {
        return fmt.Errorf("failed to create branch: %w", err)
    }

    err = r.git.Checkout(branchName)
    if err != nil {
       return fmt.Errorf("failed to checkout branch: %w", err)
    }


    return nil
}
