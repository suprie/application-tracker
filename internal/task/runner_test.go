package task

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func waitFor(t *testing.T, r *Runner, id string, want State) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		res, ok := r.Get(id)
		if ok && res.State == want {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("task %s did not reach state %s", id, want)
}

func TestRunner_Success(t *testing.T) {
	r := New(1)

	id := r.Submit("test", func(ctx context.Context) (any, error) {
		return map[string]any{"jd_id": 42}, nil
	})

	waitFor(t, r, id, StateDone)
	res, ok := r.Get(id)
	if !ok {
		t.Fatal("task not found after completion")
	}
	if res.State != StateDone {
		t.Errorf("state: want done, got %s", res.State)
	}
	m, _ := res.Result.(map[string]any)
	if m["jd_id"] != 42 {
		t.Errorf("result: got %v", res.Result)
	}
	if res.Error != "" {
		t.Errorf("unexpected error: %q", res.Error)
	}
	if res.StartedAt == nil || res.FinishedAt == nil {
		t.Error("expected timestamps to be set")
	}
}

func TestRunner_Failure(t *testing.T) {
	r := New(1)

	id := r.Submit("test", func(ctx context.Context) (any, error) {
		return nil, errors.New("boom")
	})

	waitFor(t, r, id, StateFailed)
	res, _ := r.Get(id)
	if res.State != StateFailed {
		t.Errorf("state: want failed, got %s", res.State)
	}
	if res.Error != "boom" {
		t.Errorf("error: want boom, got %q", res.Error)
	}
	if res.Result != nil {
		t.Errorf("result should be nil on failure, got %v", res.Result)
	}
}

func TestRunner_CancelPropagatesContext(t *testing.T) {
	r := New(1)

	cancelled := make(chan struct{})
	id := r.Submit("test", func(ctx context.Context) (any, error) {
		<-ctx.Done()
		close(cancelled)
		return nil, ctx.Err()
	})

	if !r.Cancel(id) {
		t.Fatal("cancel returned false for known task")
	}

	select {
	case <-cancelled:
	case <-time.After(time.Second):
		t.Fatal("context was not cancelled")
	}

	waitFor(t, r, id, StateFailed)
	res, _ := r.Get(id)
	if res.State != StateFailed {
		t.Errorf("state: want failed after cancel, got %s", res.State)
	}
}

func TestRunner_GetUnknown(t *testing.T) {
	r := New(1)
	if _, ok := r.Get("nope"); ok {
		t.Error("expected ok=false for unknown task")
	}
	if r.Cancel("nope") {
		t.Error("expected cancel=false for unknown task")
	}
}

func TestRunner_Concurrency(t *testing.T) {
	r := New(3)

	var mu sync.Mutex
	completed := 0
	ids := make([]string, 6)
	for i := 0; i < 6; i++ {
		ids[i] = r.Submit("test", func(ctx context.Context) (any, error) {
			time.Sleep(10 * time.Millisecond)
			mu.Lock()
			completed++
			mu.Unlock()
			return completed, nil
		})
	}

	for _, id := range ids {
		waitFor(t, r, id, StateDone)
	}
	if completed != 6 {
		t.Errorf("want 6 completed, got %d", completed)
	}
}
