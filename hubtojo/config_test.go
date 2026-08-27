package main

import (
	"context"
	"errors"
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
