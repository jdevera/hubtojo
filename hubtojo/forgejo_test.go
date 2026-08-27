package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-github/v63/github"
)

func TestForgejoMirrorDryRunTreatsNotFoundAsMissing(t *testing.T) {
	server := newForgejoTestServer(t, http.StatusNotFound)
	defer server.Close()

	result, err := ForgejoMirror(context.Background(), testGithubRepository(), Config{
		ForgejoUrl:      server.URL,
		ForgejoUsername: "owner",
		DryRun:          true,
	})
	if err != nil {
		t.Fatalf("mirror repository: %v", err)
	}
	if result != WouldCreate {
		t.Fatalf("result = %v, want %v", result, WouldCreate)
	}
}

func TestForgejoMirrorDryRunReturnsLookupErrors(t *testing.T) {
	server := newForgejoTestServer(t, http.StatusInternalServerError)
	defer server.Close()

	result, err := ForgejoMirror(context.Background(), testGithubRepository(), Config{
		ForgejoUrl:      server.URL,
		ForgejoUsername: "owner",
		DryRun:          true,
	})
	if err == nil {
		t.Fatal("expected repository lookup error")
	}
	if result != Failed {
		t.Fatalf("result = %v, want %v", result, Failed)
	}
}

func newForgejoTestServer(t *testing.T, repoStatus int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/version":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"version":"11.0.10"}`)
		case "/api/v1/repos/owner/repository":
			w.WriteHeader(repoStatus)
			fmt.Fprint(w, `{"message":"test response"}`)
		default:
			http.NotFound(w, r)
		}
	}))
}

func testGithubRepository() *github.Repository {
	return &github.Repository{
		Name:     github.String("repository"),
		FullName: github.String("source/repository"),
	}
}
