package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/go-github/v63/github"
)

func TestForgejoMirrorMigratesMissingRepository(t *testing.T) {
	type migrateRequest struct {
		RepoName    string `json:"repo_name"`
		RepoOwner   string `json:"repo_owner"`
		CloneAddr   string `json:"clone_addr"`
		AuthToken   string `json:"auth_token"`
		Mirror      bool   `json:"mirror"`
		Private     bool   `json:"private"`
		Description string `json:"description"`
	}
	migrations := make(chan migrateRequest, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Header.Get("Authorization") != "token forgejo-token" {
			http.Error(w, `{"message":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/version":
			fmt.Fprint(w, `{"version":"11.0.10"}`)
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/repos/owner/repository":
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"message":"not found"}`)
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/repos/migrate":
			var request migrateRequest
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				http.Error(w, `{"message":"invalid request"}`, http.StatusBadRequest)
				return
			}
			migrations <- request
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, `{"full_name":"owner/repository"}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	githubToken := "github-token"
	repository := testGithubRepository()
	repository.Private = github.Bool(true)
	repository.Description = github.String("Repository description")
	result, err := ForgejoMirror(context.Background(), repository, Config{
		ForgejoUrl:      server.URL,
		ForgejoToken:    "forgejo-token",
		ForgejoUsername: "owner",
		GithubToken:     &githubToken,
	})
	if err != nil {
		t.Fatalf("mirror repository: %v", err)
	}
	if result != Created {
		t.Fatalf("result = %v, want %v", result, Created)
	}

	select {
	case request := <-migrations:
		want := migrateRequest{
			RepoName:    "repository",
			RepoOwner:   "owner",
			CloneAddr:   "https://github.test/source/repository.git",
			AuthToken:   "github-token",
			Mirror:      true,
			Private:     true,
			Description: "Repository description",
		}
		if request != want {
			t.Fatalf("migration request = %+v, want %+v", request, want)
		}
	case <-time.After(time.Second):
		t.Fatal("Forgejo did not receive a migration request")
	}
}

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
		CloneURL: github.String("https://github.test/source/repository.git"),
		Private:  github.Bool(false),
		Fork:     github.Bool(false),
	}
}
