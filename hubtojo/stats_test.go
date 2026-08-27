package main

import (
	"slices"
	"testing"
	"time"
)

func TestRunStatsRecordsRepositoryOutcomes(t *testing.T) {
	var stats RunStats
	results := []RepoSyncResult{
		{Name: "source/created", Result: Created},
		{Name: "source/skipped", Result: Skipped},
		{Name: "source/dry-run", Result: WouldCreate},
		{Name: "source/failed", Result: Failed, Error: "migration failed"},
	}
	for _, result := range results {
		stats.record(result)
	}

	if stats.Created != 1 || stats.Skipped != 1 || stats.WouldCreate != 1 || stats.Failed != 1 {
		t.Fatalf("unexpected repository counts: %+v", stats)
	}
	if !slices.Equal(stats.CreatedRepositories, []string{"source/created"}) {
		t.Fatalf("created repositories = %v", stats.CreatedRepositories)
	}
	if !slices.Equal(stats.WouldCreateRepos, []string{"source/dry-run"}) {
		t.Fatalf("dry-run repositories = %v", stats.WouldCreateRepos)
	}
	wantFailures := []RepoFailure{{Name: "source/failed", Error: "migration failed"}}
	if !slices.Equal(stats.FailedRepositories, wantFailures) {
		t.Fatalf("failed repositories = %v", stats.FailedRepositories)
	}
}

func TestStatsSnapshotsDoNotExposeStoredSlices(t *testing.T) {
	store := NewStatsStore("test", 60)
	startedAt := time.Now()
	finishedAt := startedAt.Add(time.Second)
	store.FinishRun(RunStats{
		StartedAt:           startedAt,
		CreatedRepositories: []string{"source/created"},
		WouldCreateRepos:    []string{"source/dry-run"},
		FailedRepositories:  []RepoFailure{{Name: "source/failed"}},
	}, finishedAt)

	snapshot := store.Snapshot()
	snapshot.LastRun.CreatedRepositories[0] = "changed"
	snapshot.LastRun.WouldCreateRepos[0] = "changed"
	snapshot.LastRun.FailedRepositories[0].Name = "changed"

	stored := store.Snapshot().LastRun
	if stored.CreatedRepositories[0] != "source/created" {
		t.Fatalf("stored created repositories changed: %v", stored.CreatedRepositories)
	}
	if stored.WouldCreateRepos[0] != "source/dry-run" {
		t.Fatalf("stored dry-run repositories changed: %v", stored.WouldCreateRepos)
	}
	if stored.FailedRepositories[0].Name != "source/failed" {
		t.Fatalf("stored failures changed: %v", stored.FailedRepositories)
	}
}
