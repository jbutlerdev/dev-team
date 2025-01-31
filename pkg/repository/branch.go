package repository

import (
	"fmt"
	"log"
	"strings"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func (r *Repository) createBranch(issue *Issue) error {
    repo, err := git.PlainOpen(r.Path)
    if err != nil {
        return fmt.Errorf("error opening repository: %w", err)
    }

	// Get the main branch
	mainBranch, err := repo.Reference(plumbing.ReferenceName("refs/heads/main"), true)
	if err != nil {
		return fmt.Errorf("error getting main branch: %w", err)
	}

	// Get the commit object for the main branch
	commit, err := repo.CommitObject(mainBranch.Hash())
	if err != nil {
		return fmt.Errorf("error getting commit object: %w", err)
	}

    branchName := issue.Title
    branchName = strings.ToLower(branchName)
    branchName = strings.ReplaceAll(branchName, " ", "-")
    if len(branchName) > 100 {
        branchName = branchName[:100]
    }

    // Create the new branch
    newBranch := plumbing.ReferenceName("refs/heads/" + branchName)
    _, err = repo.CreateBranch(&git.CreateBranchOptions{
        Name:   branchName,
        From:   commit,
    })
    if err != nil {
        return fmt.Errorf("error creating branch: %w", err)
    }

    // Checkout the new branch
    w, err := repo.Worktree()
    if err != nil {
        return fmt.Errorf("error getting worktree: %w", err)
    }
    err = w.Checkout(&git.CheckoutOptions{Branch: newBranch, Create: false})
    if err != nil {
        return fmt.Errorf("error checking out branch: %w", err)
    }

    log.Printf("Created and checked out branch %s\n", branchName)

	r.State.CurrentBranch = branchName

    return nil
}
