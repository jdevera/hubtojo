package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/google/go-github/v63/github"
)

// Version of the application. It will be set during the build process.
var Version = "dev"

func MirrorWorker(ctx context.Context, id int, wg *sync.WaitGroup, repos <-chan *github.Repository, stats chan<- RepoSyncResult, config Config) {
	defer wg.Done()
	log.Printf("[Worker %d] Starting\n", id)
	ctx = context.WithValue(ctx, "worker_id", id)
	for repo := range repos {
		log.Printf("[Worker %d] Processing repository %s\n", id, *repo.FullName)
		res, err := GiteaMirror(ctx, repo, config)
		result := RepoSyncResult{
			Name:   *repo.FullName,
			Result: res,
		}
		if err != nil {
			log.Printf("[Worker %d] Error mirroring repository %s: %s\n", id, *repo.FullName, err)
			result.Error = err.Error()
		}
		stats <- result
	}
	log.Printf("[Worker %d] Done\n", id)
}

func SyncRepoList(ctx context.Context, config Config) (RunStats, error) {
	repos, err := GetGithubRepos(ctx, config)
	if err != nil {
		return RunStats{}, err
	}

	log.Printf("Found %d repositories\n", len(repos))
	for _, repo := range repos {
		log.Printf("Repository -> name: %v, private=%v, fork=%v\n", *repo.FullName, *repo.Private, *repo.Fork)
	}

	repoChan := make(chan *github.Repository, len(repos))
	statsChan := make(chan RepoSyncResult, len(repos))
	var wg sync.WaitGroup

	resultsStats := RunStats{
		TotalRead: len(repos),
	}

	for workerId := 0; workerId < config.NumWorkers; workerId++ {
		wg.Add(1)
		go MirrorWorker(ctx, workerId, &wg, repoChan, statsChan, config)
	}

	for _, repo := range repos {
		repoChan <- repo
	}
	close(repoChan)

	wg.Wait()

	close(statsChan)
	for mirrorResult := range statsChan {
		resultsStats.record(mirrorResult)
	}

	return resultsStats, nil
}

func runEvery(interval time.Duration, store *StatsStore, f func(int) RunStats) {
	runCount := 1
	for {
		startTime := time.Now()
		store.StartRun(runCount, startTime)
		stats := f(runCount)
		stats.RunCount = runCount
		stats.StartedAt = startTime
		elapsed := time.Since(startTime)
		store.FinishRun(stats, time.Now())
		if interval <= 0 {
			store.ClearNextRun()
			return
		}
		nextRun := interval - elapsed
		if nextRun < 0 {
			log.Printf("Operation took longer than the interval: %s\n", elapsed)
			nextRun = 0
		}
		nextRunAt := time.Now().Add(nextRun)
		store.SetNextRun(nextRunAt)
		log.Printf("Next run in ~%s\n", nextRun.Round(time.Second))
		if nextRun > 0 {
			time.Sleep(nextRun)
		}
		runCount++
	}
}

func main() {
	log.SetFlags(0)
	config, err := MakeConfigFromEnv()
	if err != nil {
		log.Fatalf("HubToTea version: %s\nConfig error: %s\n", Version, err)
	}

	statsStore := NewStatsStore(Version, config.SyncInterval)
	if _, err := StartWebServer(config.WebAddr, statsStore); err != nil {
		log.Fatalf("Web server error: %s\n", err)
	}

	runEvery(time.Duration(config.SyncInterval)*time.Second,
		statsStore,
		func(runCount int) RunStats {
			log.Println("--------------------------------------------------")
			log.Printf("HubToTea version: %s\n", Version)
			log.Printf("Run #%d\n", runCount)
			config.log()
			log.Println("--------------------------------------------------")

			runStats, err := SyncRepoList(context.Background(), config)
			log.Printf("--------------------------------------------------\n")
			log.Printf("Results:\n")
			if err != nil {
				log.Printf("  Error: %s\n", err.Error())
				log.Printf("--------------------------------------------------\n")
				runStats.Status = "error"
				runStats.Error = err.Error()
				return runStats
			}
			runStats.Status = "success"
			if runStats.Failed > 0 {
				runStats.Status = "completed_with_errors"
			}
			log.Printf("  Total Read: %d\n", runStats.TotalRead)
			log.Printf("  Created: %d\n", runStats.Created)
			log.Printf("  Skipped: %d\n", runStats.Skipped)
			log.Printf("  WouldCreate: %d\n", runStats.WouldCreate)
			log.Printf("  Failed: %d\n", runStats.Failed)
			log.Printf("--------------------------------------------------\n")
			return runStats
		})
}
