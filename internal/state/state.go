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
	Repositories map[string]*repository.Repository `json:"repositories"`
    TrackedIssues  map[int64]repository.Issue          `json:"tracked_issues"`
	Settings     settings.Settings                 `json:"settings"`
	Scheduler    *scheduler.Scheduler
	Mu           sync.RWMutex
	GenAI        *genai.Provider
}

func (s *AppState) UpdateTrackedIssues(issues map[int64]repository.Issue) {
    s.Mu.Lock()
    defer s.Mu.Unlock()
    if s.TrackedIssues == nil {
        s.TrackedIssues = make(map[int64]repository.Issue)
    }

    for id, issue := range issues {
        if issue.State == "closed" {
            delete(s.TrackedIssues, id)
        } else {
            s.TrackedIssues[id] = issue
        }
    }
}
