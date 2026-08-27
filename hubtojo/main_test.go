package main

import (
	"context"
	"testing"
	"time"
)

func TestRunEveryRunsOnceWhenIntervalIsZero(t *testing.T) {
	store := NewStatsStore("test", 0)
	calls := 0

	runEvery(context.Background(), 0, store, func(_ context.Context, runCount int) RunStats {
		calls++
		return RunStats{
			Status:    "success",
			TotalRead: 3,
		}
	})

	if calls != 1 {
		t.Fatalf("runs = %d, want 1", calls)
	}
	snapshot := store.Snapshot()
	if snapshot.CurrentRun != nil {
		t.Fatal("current run is still set after completion")
	}
	if snapshot.NextRunAt != nil {
		t.Fatal("next run is set for a one-shot synchronization")
	}
	if snapshot.LastRun == nil {
		t.Fatal("last run is not set")
	}
	if snapshot.LastRun.RunCount != 1 {
		t.Fatalf("run count = %d, want 1", snapshot.LastRun.RunCount)
	}
	if snapshot.LastRun.TotalRead != 3 {
		t.Fatalf("total read = %d, want 3", snapshot.LastRun.TotalRead)
	}
	if snapshot.LastRun.StartedAt.IsZero() || snapshot.LastRun.FinishedAt == nil {
		t.Fatal("run timestamps are incomplete")
	}
}

func TestRunEveryCancellationInterruptsWait(t *testing.T) {
	store := NewStatsStore("test", 3600)
	ctx, cancel := context.WithCancel(context.Background())
	runStarted := make(chan struct{})
	done := make(chan struct{})

	go func() {
		defer close(done)
		runEvery(ctx, time.Hour, store, func(_ context.Context, _ int) RunStats {
			close(runStarted)
			return RunStats{Status: "success"}
		})
	}()

	select {
	case <-runStarted:
	case <-time.After(time.Second):
		t.Fatal("synchronization did not start")
	}
	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("scheduler did not stop after cancellation")
	}

	snapshot := store.Snapshot()
	if snapshot.LastRun == nil {
		t.Fatal("completed run was not recorded")
	}
	if snapshot.NextRunAt != nil {
		t.Fatal("next run remains set after cancellation")
	}
}
