package github

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-git/go-git/v5"
	githubv60 "github.com/google/go-github/v60/github"
	"golang.org/x/oauth2"
)

type GitHubPRInput struct {
	Title               string `json:"title"`
	Description         string `json:"description"`
	Branch              string `json:"branch"`
	Base                string `json:"base"`
	Draft               bool   `json:"draft"`
	MaintainerCanModify bool   `json:"maintainer_can_modify"`
}

type GitHubPRResponse struct {
	Number int `json:"number"`
}

type Repository struct {
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	CloneURL    string `json:"clone_url"`
	SSHURL      string `json:"ssh_url"`
}

type Issue struct {
	Number    int    `json:"number"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	State     string `json:"state"`
	HTMLURL   string `json:"html_url"`
	SourceURL string `json:"source_url"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type PullRequest struct {
	Number          int      `json:"number"`
	Title           string   `json:"title"`
	Body            string   `json:"body"`
	State           string   `json:"state"`
	HTMLURL         string   `json:"html_url"`
	Labels          []string `json:"labels"`
	IssueURL        string   `json:"issue_url"`
	CreatedAt       string   `json:"created_at"`
	UpdatedAt       string   `json:"updated_at"`
	LinkedIssueURLs []string `json:"linked_issue_urls"`
}

type IssueEvent struct {
	Event     string `json:"event"`
	CreatedAt string `json:"created_at"`
	PRNumber  int    `json:"pr_number,omitempty"`
	PRURL     string `json:"pr_url,omitempty"`
}

var re = regexp.MustCompile(`<!--(.*?)-->`)

func newGitHubClient(ctx context.Context, token string) *githubv60.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	return githubv60.NewClient(tc)
}

func CreateDraftPR(path string, githubToken string, input GitHubPRInput) error {
	ctx := context.Background()
	client := newGitHubClient(ctx, githubToken)

	repo, err := git.PlainOpen(path)
	if err != nil {
		return err
	}

	remote, err := repo.Remote("origin")
	if err != nil {
		return fmt.Errorf("error getting remote: %v", err)
	}

	remoteURL := remote.Config().URLs[0]
	var owner, repoName string
	if strings.Contains(remoteURL, "git@github.com:") {
		parts := strings.Split(strings.TrimPrefix(remoteURL, "git@github.com:"), "/")
		owner = parts[0]
		repoName = strings.TrimSuffix(parts[1], ".git")
	} else {
		parts := strings.Split(strings.TrimPrefix(remoteURL, "https://github.com/"), "/")
		owner = parts[0]
		repoName = strings.TrimSuffix(parts[1], ".git")
	}

	if githubToken == "" {
		return fmt.Errorf("GitHub token not provided in settings")
	}

	newPR := &githubv60.NewPullRequest{
		Title:               githubv60.String(input.Title),
		Head:                githubv60.String(input.Branch),
		Base:                githubv60.String(input.Base),
		Body:                githubv60.String(input.Description),
		Draft:               githubv60.Bool(input.Draft),
		MaintainerCanModify: githubv60.Bool(input.MaintainerCanModify),
	}

	pr, _, err := client.PullRequests.Create(ctx, owner, repoName, newPR)
	if err != nil {
		return fmt.Errorf("error creating PR: %v", err)
	}

	prLink := pr.GetHTMLURL()
	log.Printf("PR created successfully: %s", prLink)

	return nil
}

func FetchRepositories(githubToken string) ([]Repository, error) {
	if githubToken == "" {
		return nil, fmt.Errorf("GitHub token not provided in settings")
	}

	ctx := context.Background()
	client := newGitHubClient(ctx, githubToken)

	opt := &githubv60.RepositoryListByAuthenticatedUserOptions{
		Sort: "updated",
		ListOptions: githubv60.ListOptions{
			PerPage: 100,
		},
	}

	repos, _, err := client.Repositories.ListByAuthenticatedUser(ctx, opt)
	if err != nil {
		return nil, fmt.Errorf("error fetching repositories: %v", err)
	}

	var result []Repository
	for _, repo := range repos {
		result = append(result, Repository{
			Name:        repo.GetName(),
			FullName:    repo.GetFullName(),
			Description: repo.GetDescription(),
			CloneURL:    repo.GetCloneURL(),
			SSHURL:      repo.GetSSHURL(),
		})
	}

	return result, nil
}

func FetchIssues(remotePath, label, githubToken string) ([]Issue, error) {
	if githubToken == "" {
		return nil, fmt.Errorf("GitHub token not provided in settings")
	}

	ctx := context.Background()
	client := newGitHubClient(ctx, githubToken)

	// Extract owner and repo from remote path
	// Expected format: owner/repo
	parts := strings.Split(remotePath, "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid remote path format")
	}
	owner := parts[0]
	repo := parts[1]

	opt := &githubv60.IssueListByRepoOptions{
		Labels: []string{label},
		State:  "open",
		ListOptions: githubv60.ListOptions{
			PerPage: 100,
		},
	}

	ghIssues, _, err := client.Issues.ListByRepo(ctx, owner, repo, opt)
	if err != nil {
		log.Printf("Error fetching issues: %v, request: %v", err, remotePath)
		return nil, fmt.Errorf("error fetching issues: %v", err)
	}

	var issues []Issue
	for _, issue := range ghIssues {
		i := Issue{
			Number:    issue.GetNumber(),
			Title:     issue.GetTitle(),
			Body:      issue.GetBody(),
			State:     issue.GetState(),
			HTMLURL:   issue.GetHTMLURL(),
			SourceURL: issue.GetHTMLURL(),
			CreatedAt: issue.GetCreatedAt().String(),
			UpdatedAt: issue.GetUpdatedAt().String(),
		}
		issues = append(issues, i)
	}

	return issues, nil
}

func FetchPullRequests(remotePath, label, githubToken string) ([]PullRequest, error) {
	if githubToken == "" {
		return nil, fmt.Errorf("GitHub token not provided in settings")
	}

	ctx := context.Background()
	client := newGitHubClient(ctx, githubToken)

	parts := strings.Split(remotePath, "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid remote path format")
	}
	owner := parts[0]
	repo := parts[1]

	opt := &githubv60.PullRequestListOptions{
		State: "open",
		ListOptions: githubv60.ListOptions{
			PerPage: 100,
		},
	}

	ghPullRequests, _, err := client.PullRequests.List(ctx, owner, repo, opt)
	if err != nil {
		log.Printf("Error fetching pull requests: %v, request: %v", err, remotePath)
		return nil, fmt.Errorf("error fetching pull requests: %v", err)
	}

	var pullRequests []PullRequest
	for _, pullRequest := range ghPullRequests {
		pr := PullRequest{
			Number:          pullRequest.GetNumber(),
			Title:           pullRequest.GetTitle(),
			Body:            pullRequest.GetBody(),
			State:           pullRequest.GetState(),
			Labels:          make([]string, 0, len(pullRequest.Labels)),
			HTMLURL:         pullRequest.GetHTMLURL(),
			IssueURL:        pullRequest.GetIssueURL(),
			CreatedAt:       pullRequest.GetCreatedAt().String(),
			UpdatedAt:       pullRequest.GetUpdatedAt().String(),
			LinkedIssueURLs: getLinkedIssueURLs(pullRequest.GetBody()),
		}
		for i, label := range pullRequest.Labels {
			pr.Labels[i] = label.GetName()
		}
		pullRequests = append(pullRequests, pr)
	}

	return pullRequests, nil
}

func getLinkedIssueURLs(body string) []string {
	// URLs are in HTML comments
	matches := re.FindAllString(body, -1)
	urls := make([]string, len(matches))
	for i, match := range matches {
		match = strings.TrimPrefix(match, "<!--")
		match = strings.TrimSuffix(match, "-->")
		urls[i] = match
	}
	return urls
}

// AddCommitToPullRequest adds a commit to a pull request.
func AddCommitToPullRequest(githubToken string, owner string, repo string, pullRequestNumber int, commitMessage string, content string, path string, branch string) error {
	if githubToken == "" {
		return fmt.Errorf("GitHub token not provided in settings")
	}

	ctx := context.Background()
	client := newGitHubClient(ctx, githubToken)

	// Get the pull request.
	_, _, err := client.PullRequests.Get(ctx, owner, repo, pullRequestNumber)
	if err != nil {
		return fmt.Errorf("error getting pull request: %v", err)
	}

	// Get the branch's reference.
	reference, _, err := client.Git.GetRef(ctx, owner, repo, "refs/heads/"+branch)
	if err != nil {
		return fmt.Errorf("error getting reference: %v", err)
	}

	// Get the commit.
	commit, _, err := client.Git.GetCommit(ctx, owner, repo, reference.GetObject().GetSHA())
	if err != nil {
		return fmt.Errorf("error getting commit: %v", err)
	}

	// Create a tree.

	// Create a blob.
	blob, _, err := client.Git.CreateBlob(ctx, owner, repo, &githubv60.Blob{Content: githubv60.String(content), Encoding: githubv60.String("utf-8")})
	if err != nil {
		return fmt.Errorf("error getting blob: %v", err)
	}

	// Create a tree entry.
	treeEntry := githubv60.TreeEntry{Path: githubv60.String(path), Type: githubv60.String("blob"), SHA: blob.SHA, Mode: githubv60.String("100644")}

	entries := []*githubv60.TreeEntry{&treeEntry}

	newTree, _, err := client.Git.CreateTree(ctx, owner, repo, commit.GetTree().GetSHA(), entries)
	if err != nil {
		return fmt.Errorf("error creating tree: %v", err)
	}

	// Create a commit.

	newCommit := &githubv60.Commit{
		Message: githubv60.String(commitMessage),
		Tree:    newTree,
		Parents: []*githubv60.Commit{&githubv60.Commit{SHA: commit.SHA}},
	}

	commitResponse, _, err := client.Git.CreateCommit(ctx, owner, repo, newCommit, &githubv60.CreateCommitOptions{})
	if err != nil {
		return fmt.Errorf("error creating commit: %v", err)
	}

	// Update the reference.

	newReference := &githubv60.Reference{
		Ref: githubv60.String("refs/heads/" + branch),
		Object: &githubv60.GitObject{
			Type: githubv60.String("commit"),
			SHA:  commitResponse.SHA,
		},
	}

	_, _, err = client.Git.UpdateRef(ctx, owner, repo, newReference, false)
	if err != nil {
		return fmt.Errorf("error updating reference: %v", err)
	}

	return nil
}

// ValidateWebhook validates the webhook signature.
func ValidateWebhook(r *http.Request, secret string) ([]byte, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	signature := r.Header.Get("X-Hub-Signature-256")
	if signature == "" {
		return nil, errors.New("X-Hub-Signature-256 header is missing")
	}

	hmacValue := hmac.New(sha256.New, []byte(secret))
	hmacValue.Write(body)
	expectedSignature := "sha256=" + hex.EncodeToString(hmacValue.Sum(nil))

	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return nil, errors.New("invalid signature")
	}

	return body, nil
}
