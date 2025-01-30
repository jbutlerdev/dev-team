package state

import (
	"dev-team/internal/scheduler"
	"dev-team/internal/settings"
	"dev-team/pkg/repository"
	"genai"
	"sync"
)

var State *AppState

type AppState struct {
	Repositories  map[string]*repository.Repository `json:"repositories"`
	Settings      settings.Settings                 `json:"settings"`
	Scheduler     *scheduler.Scheduler
	Mu            sync.RWMutex
	GenAI         *genai.Provider
	TrackedIssues map[int]*repository.Issue `json:"tracked_issues"`
}

func (a *AppState) AddTrackedIssue(issue *repository.Issue) {
	a.Mu.Lock()
	defer a.Mu.Unlock()
	if a.TrackedIssues == nil {
		a.TrackedIssues = make(map[int]*repository.Issue)
	}
	a.TrackedIssues[issue.GetID()] = issue
}

func (a *AppState) RemoveTrackedIssue(issue *repository.Issue) {
	a.Mu.Lock()
	defer a.Mu.Unlock()
	delete(a.TrackedIssues, issue.GetID())
}

func (a *AppState) UpdateTrackedIssue(issue *repository.Issue) {
    a.Mu.Lock()
    defer a.Mu.Unlock()
    if _, ok := a.TrackedIssues[issue.GetID()]; ok {
        a.TrackedIssues[issue.GetID()] = issue
    }
}
