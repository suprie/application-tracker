// Package task provides an in-memory, bounded-concurrency job runner for the
// slow LLM-backed operations exposed by the web server. Jobs are identified by
// an opaque ID and polled via Get. State is kept only in memory: in-flight jobs
// are lost if the process restarts, which is acceptable for a single-user
// local tool.
package task

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

// State is the lifecycle state of a Task.
type State string

const (
	StatePending State = "pending"
	StateRunning State = "running"
	StateDone    State = "done"
	StateFailed  State = "failed"
)

// Task is the internal, mutable record of a job.
type Task struct {
	ID         string
	Kind       string
	State      State
	Result     any
	Error      string
	StartedAt  *time.Time
	FinishedAt *time.Time
	cancel     context.CancelFunc
}

// Result is the JSON-friendly, value snapshot returned to API callers.
type Result struct {
	ID         string     `json:"id"`
	Kind       string     `json:"kind"`
	State      State      `json:"state"`
	Result     any        `json:"result,omitempty"`
	Error      string     `json:"error,omitempty"`
	StartedAt  *time.Time `json:"started_at,omitempty"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
}

// Runner dispatches jobs to a fixed pool of workers.
type Runner struct {
	mu    sync.RWMutex
	tasks map[string]*Task
	work  chan job
}

type job struct {
	id  string
	ctx context.Context
	fn  func(ctx context.Context) (any, error)
}

// New returns a Runner with the given worker concurrency (min 1).
func New(concurrency int) *Runner {
	if concurrency < 1 {
		concurrency = 1
	}
	r := &Runner{
		tasks: make(map[string]*Task),
		work:  make(chan job, 64),
	}
	for i := 0; i < concurrency; i++ {
		go r.worker()
	}
	return r
}

func (r *Runner) worker() {
	for j := range r.work {
		r.runJob(j)
	}
}

func (r *Runner) runJob(j job) {
	r.mu.Lock()
	t := r.tasks[j.id]
	t.State = StateRunning
	now := time.Now()
	t.StartedAt = &now
	r.mu.Unlock()

	result, err := j.fn(j.ctx)

	r.mu.Lock()
	t.Result = result
	fin := time.Now()
	t.FinishedAt = &fin
	if err != nil {
		t.State = StateFailed
		t.Error = err.Error()
	} else {
		t.State = StateDone
	}
	r.mu.Unlock()
}

// Submit queues a job and returns its ID immediately. fn runs on a worker,
// receiving a cancellable context.
func (r *Runner) Submit(kind string, fn func(ctx context.Context) (any, error)) string {
	id := uuid.NewString()
	ctx, cancel := context.WithCancel(context.Background())
	r.mu.Lock()
	r.tasks[id] = &Task{ID: id, Kind: kind, State: StatePending, cancel: cancel}
	r.mu.Unlock()
	r.work <- job{id: id, ctx: ctx, fn: fn}
	return id
}

// Get returns a snapshot of a task, or ok=false if unknown.
func (r *Runner) Get(id string) (*Result, bool) {
	r.mu.RLock()
	t, ok := r.tasks[id]
	r.mu.RUnlock()
	if !ok {
		return nil, false
	}
	return &Result{
		ID:         t.ID,
		Kind:       t.Kind,
		State:      t.State,
		Result:     t.Result,
		Error:      t.Error,
		StartedAt:  t.StartedAt,
		FinishedAt: t.FinishedAt,
	}, true
}

// Cancel requests cancellation of a running task. Returns ok=false if unknown.
// Best-effort: fn must respect its context to actually stop.
func (r *Runner) Cancel(id string) bool {
	r.mu.RLock()
	t, ok := r.tasks[id]
	r.mu.RUnlock()
	if !ok {
		return false
	}
	if t.cancel != nil {
		t.cancel()
	}
	return true
}
