package local

import "github.com/jbutlerdev/dev-team/pkg/remote/github"

type Provider struct {
	Path string
}

func NewProvider(path string) *Provider {
	return &Provider{
		Path: path,
	}
}

func (p *Provider) CreateDraftPR(path string, input github.GitHubPRInput) error {
	return nil
}

func (p *Provider) FetchRepositories() ([]github.Repository, error) {
	return nil, nil
}

func (p *Provider) FetchIssues(remotePath, label string) ([]github.Issue, error) {
	return nil, nil
}

func (p *Provider) FetchPullRequests(remotePath, label string) ([]github.PullRequest, error) {
	return nil, nil
}

func (p *Provider) FetchDiffs(owner, repo string, resourceID int) (string, error) {
	return "", nil
}

func (p *Provider) FetchComments(owner, repo string, prNumber int) ([]github.Comment, error) {
	return nil, nil
}

func (p *Provider) AddCommentReaction(repoPath, reaction string, commentID int64) error {
	return nil
}
