package scheduler

import (
	"testing"
	"time"
)

func TestScheduler_AddTask(t *testing.T) {
	s := NewScheduler()
	s.Start()
	defer s.Stop()

	var taskRun bool

	err := s.AddTask("test_task", "@every 1s", func() {
		taskRun = true
	})

	if err != nil {
		t.Fatalf("Error adding task: %v", err)
	}

	time.Sleep(2 * time.Second)

	if !taskRun {
		t.Errorf("Task did not run")
	}

	s.RemoveTask("test_task")
}

func TestScheduler_RemoveTask(t *testing.T) {
	s := NewScheduler()
	s.Start()
	defer s.Stop()

	var taskRun bool

	err := s.AddTask("test_task", "@every 1s", func() {
		taskRun = true
	})

	if err != nil {
		t.Fatalf("Error adding task: %v", err)
	}

	s.RemoveTask("test_task")

	time.Sleep(2 * time.Second)

	if taskRun {
		t.Errorf("Task should not have run after removal")
	}
}
