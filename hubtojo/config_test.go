package main

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestRunContextUsesConfiguredTimeout(t *testing.T) {
	config := Config{RunTimeout: 10 * time.Millisecond}
	ctx, cancel := config.withRunTimeout(context.Background())
	defer cancel()

	select {
	case <-ctx.Done():
		if !errors.Is(ctx.Err(), context.DeadlineExceeded) {
			t.Fatalf("context error = %v, want deadline exceeded", ctx.Err())
		}
	case <-time.After(time.Second):
		t.Fatal("run context did not reach its deadline")
	}
}

func TestConfigValidateRejectsUnsafeNumericValues(t *testing.T) {
	tests := []struct {
		name      string
		configure func(*Config)
		wantError string
	}{
		{
			name:      "zero workers",
			configure: func(config *Config) { config.NumWorkers = 0 },
			wantError: "HUBTOJO_NUM_WORKERS",
		},
		{
			name:      "negative sync interval",
			configure: func(config *Config) { config.SyncInterval = -1 },
			wantError: "HUBTOJO_SYNC_INTERVAL",
		},
		{
			name:      "zero run timeout",
			configure: func(config *Config) { config.RunTimeout = 0 },
			wantError: "HUBTOJO_RUN_TIMEOUT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := validConfig()
			tt.configure(&config)
			err := config.validate()
			if err == nil || !strings.Contains(err.Error(), tt.wantError) {
				t.Fatalf("validation error = %v, want %s", err, tt.wantError)
			}
		})
	}
}

func TestConfigValidateRejectsEmptyPrivateRepoToken(t *testing.T) {
	config := validConfig()
	emptyToken := ""
	config.GithubToken = &emptyToken
	config.MirrorPrivateRepos = true

	err := config.validate()
	if err == nil || !strings.Contains(err.Error(), "GITHUB_TOKEN") {
		t.Fatalf("validation error = %v, want GITHUB_TOKEN error", err)
	}
}

func validConfig() Config {
	return Config{
		GithubUsername: "source-user",
		ForgejoUrl:     "https://forgejo.example.com",
		ForgejoToken:   "token",
		NumWorkers:     5,
		SyncInterval:   3600,
		RunTimeout:     time.Hour,
	}
}
