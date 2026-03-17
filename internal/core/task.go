// Package core provides the core engine
package core

import (
	"context"
	"time"
)

// Task represents a scheduled task
type Task struct {
	ID          string
	Name        string
	Description string
	Schedule    string
	Handler     TaskHandler
	LastRun     time.Time
	NextRun     time.Time
	Enabled     bool
}

// TaskHandler is a function that handles a task
type TaskHandler func(ctx context.Context) error

// TaskScheduler manages scheduled tasks
type TaskScheduler struct {
	tasks map[string]*Task
}

// NewTaskScheduler creates a new task scheduler
func NewTaskScheduler() *TaskScheduler {
	return &TaskScheduler{
		tasks: make(map[string]*Task),
	}
}

// Register registers a new task
func (ts *TaskScheduler) Register(task *Task) {
	ts.tasks[task.ID] = task
}

// Unregister removes a task
func (ts *TaskScheduler) Unregister(id string) {
	delete(ts.tasks, id)
}

// Get retrieves a task by ID
func (ts *TaskScheduler) Get(id string) (*Task, bool) {
	task, ok := ts.tasks[id]
	return task, ok
}

// List returns all tasks
func (ts *TaskScheduler) List() []*Task {
	tasks := make([]*Task, 0, len(ts.tasks))
	for _, t := range ts.tasks {
		tasks = append(tasks, t)
	}
	return tasks
}

// Run executes a task by ID
func (ts *TaskScheduler) Run(ctx context.Context, id string) error {
	task, ok := ts.tasks[id]
	if !ok {
		return nil
	}

	if !task.Enabled {
		return nil
	}

	err := task.Handler(ctx)
	task.LastRun = time.Now()

	return err
}

// RunAll executes all enabled tasks
func (ts *TaskScheduler) RunAll(ctx context.Context) []error {
	var errs []error

	for _, task := range ts.tasks {
		if !task.Enabled {
			continue
		}

		if err := ts.Run(ctx, task.ID); err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}
