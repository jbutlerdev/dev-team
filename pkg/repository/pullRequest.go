package repository

import (
	"fmt"
	"log"

	"github.com/jbutlerdev/dev-team/pkg/remote/github"
)

type PullRequest struct {
	Number          int
	Title           string
	Body            string
	CreatedAt       string
	UpdatedAt       string
	Labels          []string
	IssueUrl        string
	LinkedIssueUrls []string
	Diff            string
	Comments        []Comment
}

type Comment struct {
	ID           int64
	Body         string
	DiffHunk     string
	HTMLURL      string
	URL          string
	UserID       int64
	Acknowledged bool
}

func (p *PullRequest) HasUnresolvedComments() bool {
	for _, comment := range p.Comments {
		if !comment.Acknowledged {
			return true
		}
	}
	return false
}

func (p *PullRequest) FirstUnresolvedComment() Comment {
	for _, comment := range p.Comments {
		if !comment.Acknowledged {
			return comment
		}
	}
	return Comment{}
}

func (r *Repository) GetPullRequests() []*PullRequest {
	pullRequests := make([]*PullRequest, 0, len(r.PullRequests))
	for _, pullRequest := range r.PullRequests {
		pullRequests = append(pullRequests, pullRequest)
	}
	return pullRequests
}

func (r *Repository) UpdatePullRequests(token string) error {
	if r.RemotePath == "" {
		return fmt.Errorf("repository remote path is not set")
	}
	pullRequests, err := r.Remote.FetchPullRequests(r.RemotePath, "dev-team")
	if err != nil {
		log.Printf("Error fetching pull requests: %v, request: %v", err, r.RemotePath)
		return err
	}
	// reset tracked pull requests
	r.PullRequests = make(map[int]*PullRequest)
	for _, pullRequest := range pullRequests {
		r.PullRequests[pullRequest.Number] = ghPullRequestToPullRequest(pullRequest)
	}
	return nil
}

func ghPullRequestToPullRequest(pullRequest github.PullRequest) *PullRequest {
	return &PullRequest{
		Number:          pullRequest.Number,
		Title:           pullRequest.Title,
		Body:            pullRequest.Body,
		CreatedAt:       pullRequest.CreatedAt,
		UpdatedAt:       pullRequest.UpdatedAt,
		Labels:          pullRequest.Labels,
		IssueUrl:        pullRequest.IssueURL,
		LinkedIssueUrls: pullRequest.LinkedIssueURLs,
		Diff:            pullRequest.Diff,
		Comments:        ghCommentsToComments(pullRequest.Comments),
	}
}

func ghCommentsToComments(comments []github.Comment) []Comment {
	pullRequestComments := make([]Comment, 0, len(comments))
	for _, comment := range comments {
		pullRequestComments = append(pullRequestComments, Comment{
			ID:           comment.ID,
			Body:         comment.Body,
			DiffHunk:     comment.DiffHunk,
			HTMLURL:      comment.HTMLURL,
			URL:          comment.URL,
			UserID:       comment.UserID,
			Acknowledged: comment.Reactions.PlusOne > 0,
		})
	}
	return pullRequestComments
}
