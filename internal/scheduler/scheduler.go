package scheduler

import (
	"dev-team/internal/state"
	"log"
	"sync"

	"github.com/robfig/cron/v3"
)

type Task struct {
	ID       cron.EntryID
	Schedule string
	Action   func()
}

type Scheduler struct {
	cron  *cron.Cron
	tasks map[string]*Task
	mu    sync.RWMutex
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		cron:  cron.New(),
		tasks: make(map[string]*Task),
	}
}

func (s *Scheduler) Start() {
	s.cron.Start()
}

func (s *Scheduler) Stop() {
	s.cron.Stop()
}

func (s *Scheduler) AddTask(key string, schedule string, action func()) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Remove existing task if it exists
	if existingTask, exists := s.tasks[key]; exists {
		s.cron.Remove(existingTask.ID)
		delete(s.tasks, key)
	}

	id, err := s.cron.AddFunc(schedule, func() {
		log.Printf("Running scheduled task for %s", key)
		action()
	})

	if err != nil {
		return err
	}

	s.tasks[key] = &Task{
		ID:       id,
		Schedule: schedule,
		Action:   action,
	}

	return nil
}

func (s *Scheduler) RemoveTask(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if task, exists := s.tasks[key]; exists {
		s.cron.Remove(task.ID)
		delete(s.tasks, key)
	}
}

func (s *Scheduler) UpdateTask(key string, schedule string, action func()) error {
	return s.AddTask(key, schedule, action)
}

func (s *Scheduler) ScheduleIssueUpdates() {
    s.AddTask("github-issue-updates", "@every 5m", func() {
        state.State.Mu.RLock()
        repos := state.State.Repositories
        token := state.State.Settings.GitHubToken
        state.State.Mu.RUnlock()

        if token == "" {
            log.Println("GitHub token not configured, skipping issue updates")
            return
        }

        for _, repo := range repos {
            err := repo.UpdateIssues(token)
            if err != nil {
                log.Printf("Error updating issues for repo %s: %v", repo.Path, err)
                continue
            }
            filteredIssues := make(map[int64]repository.Issue)
            for _, issue := range repo.Issues {
                hasDevTeamLabel := false
                for _, label := range issue.Labels {
                    if label.Name == "dev-team" {
                        hasDevTeamLabel = true
                        break
                    }
                }
                if hasDevTeamLabel {
                    filteredIssues[issue.ID] = issue
                }
            }
            state.State.UpdateTrackedIssues(filteredIssues)
        }
    })
}
