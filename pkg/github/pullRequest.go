package github

import (
	"context"
	"fmt"
    "log"
	"github.com/google/go-github/v60/github"
)


func CreatePR(path string, githubToken string, input GitHubPRInput) error {
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

	newPR := &github.NewPullRequest{
		Title:               github.String(input.Title),
		Head:                github.String(input.Branch),
		Base:                github.String(input.Base),
		Body:                github.String(input.Description),
		Draft:               github.Bool(input.Draft),
		MaintainerCanModify: github.Bool(input.MaintainerCanModify),
	}

	pr, _, err := client.PullRequests.Create(ctx, owner, repoName, newPR)
    if err != nil {
        return fmt.Errorf("error creating PR: %v", err)
    }

    _, _, err = client.Issues.AddLabelsToIssue(ctx, owner, repoName, pr.GetNumber(), []string{"dev-team"})
    if err != nil {
        return fmt.Errorf("error adding labels to PR: %v", err)
    }

	prLink := pr.GetHTMLURL()
	log.Printf("PR created successfully: %s", prLink)

	return nil
}
