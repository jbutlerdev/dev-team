package github

import (
	"context"

	"github.com/google/go-github/v60/github"
)

type Provider struct {
	Client *github.Client
	ctx    context.Context
}

func NewProvider(token string) *Provider {
	return &Provider{
		Client: newGitHubClient(context.Background(), token),
		ctx:    context.Background(),
	}
}
