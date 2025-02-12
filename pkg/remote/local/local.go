package local

import (
	"fmt"

	"github.com/jbutlerdev/dev-team/pkg/remote/types"
)

type Provider struct {
	Path         string
	Issues       map[int]*types.Issue
	PullRequests map[int]*types.PullRequest
	issueCounter int
}

func NewProvider(path string) *Provider {
	return &Provider{
		Path:         path,
		Issues:       make(map[int]*types.Issue),
		PullRequests: make(map[int]*types.PullRequest),
		issueCounter: 0,
	}
}

func (p *Provider) CreateDraftPR(path string, input types.PullRequestInput) error {
	p.issueCounter++
	p.PullRequests[p.issueCounter] = &types.PullRequest{
		Number: p.issueCounter,
		Title:  input.Title,
		Body:   input.Description,
		State:  "draft",
	}
	return nil
}

func (p *Provider) FetchRepositories() ([]types.Repository, error) {
	return nil, nil
}

func (p *Provider) FetchIssues(remotePath, label string) ([]types.Issue, error) {
	issues := make([]types.Issue, len(p.Issues))
	for i := range p.Issues {
		issues[i] = *p.Issues[i]
	}
	return issues, nil
}

func (p *Provider) FetchPullRequests(remotePath, label string) ([]types.PullRequest, error) {
	pullRequests := make([]types.PullRequest, len(p.PullRequests))
	for i := range p.PullRequests {
		pullRequests[i] = *p.PullRequests[i]
	}
	return pullRequests, nil
}

func (p *Provider) FetchDiffs(owner, repo string, resourceID int) (string, error) {
	return "", nil
}

func (p *Provider) FetchComments(owner, repo string, prNumber int) ([]types.Comment, error) {
	pr, ok := p.PullRequests[prNumber]
	if !ok {
		return nil, fmt.Errorf("pull request %d not found", prNumber)
	}
	return pr.Comments, nil
}

func (p *Provider) AddCommentReaction(repoPath, reaction string, commentID int64) error {
	for _, pr := range p.PullRequests {
		for _, comment := range pr.Comments {
			if comment.ID == commentID {
				comment.Reactions = addReactionToReactions(comment.Reactions, reaction)
			}
		}
	}
	return nil
}

func addReactionToReactions(reactions types.Reactions, reaction string) types.Reactions {
	reactions.TotalCount++
	switch reaction {
	case "+1":
		reactions.PlusOne++
	case "-1":
		reactions.MinusOne++
	case "laugh":
		reactions.Laugh++
	case "confused":
		reactions.Confused++
	case "heart":
		reactions.Heart++
	case "hooray":
		reactions.Hooray++
	case "rocket":
		reactions.Rocket++
	case "eyes":
		reactions.Eyes++
	}
	return reactions
}
