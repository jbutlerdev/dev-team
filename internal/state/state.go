package state

import (
	"dev-team/internal/scheduler"
	"dev-team/internal/settings"
	"dev-team/pkg/github"
	"dev-team/pkg/repository"
	"genai"
	"sync"
)

var State *AppState

type AppState struct {
	Repositories map[string]*repository.Repository `json:"repositories"`
	Settings     settings.Settings                 `json:"settings"`
	Scheduler    *scheduler.Scheduler
    TrackedIssues map[int]*github.Issue
	Mu           sync.RWMutex
	GenAI        *genai.Provider
}

func NewAppState() *AppState {
    return &AppState{
        Repositories: make(map[string]*repository.Repository),
        Settings:     settings.Settings{},
        Scheduler:    scheduler.NewScheduler(),
        TrackedIssues: make(map[int]*github.Issue),
        Mu:           sync.RWMutex{},
    }
}
