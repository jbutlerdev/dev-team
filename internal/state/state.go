package state

import (
	"sync"

	"github.com/jbutlerdev/dev-team/internal/scheduler"
	"github.com/jbutlerdev/dev-team/internal/settings"
	"github.com/jbutlerdev/dev-team/pkg/repository"
	"github.com/jbutlerdev/genai"
)

var State *AppState

type AppState struct {
	Repositories        map[string]*repository.Repository `json:"repositories"`
	Settings            settings.Settings                 `json:"settings"`
	Scheduler           *scheduler.Scheduler
	Mu                  sync.RWMutex
	GenAI               *genai.Provider
	TrackedPullRequests map[string]bool `json:"tracked_pull_requests"` // Map of pull request URLs to a boolean indicating if they are tracked
}

func (s *AppState) IsTracked(pullRequestURL string) bool {
	s.Mu.RLock()
	defer s.Mu.RUnlock()
	return s.TrackedPullRequests[pullRequestURL]
}

func (s *AppState) TrackPullRequest(pullRequestURL string) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	if s.TrackedPullRequests == nil {
		s.TrackedPullRequests = make(map[string]bool)
	}
	s.TrackedPullRequests[pullRequestURL] = true
}

func (s *AppState) UntrackPullRequest(pullRequestURL string) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	delete(s.TrackedPullRequests, pullRequestURL)
}
