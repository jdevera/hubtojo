package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/google/go-github/v63/github"
)

func TestPrivateRepoTokenMustMatchConfiguredUser(t *testing.T) {
	tests := []struct {
		name      string
		login     string
		wantError bool
	}{
		{name: "matching user", login: "SourceUser"},
		{name: "case insensitive match", login: "sourceuser"},
		{name: "different user", login: "another-user", wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch r.URL.Path {
				case "/user":
					fmt.Fprintf(w, `{"login":%q}`, tt.login)
				case "/user/repos":
					fmt.Fprint(w, `[]`)
				default:
					http.NotFound(w, r)
				}
			}))
			defer server.Close()

			client := github.NewClient(server.Client())
			client.BaseURL, _ = url.Parse(server.URL + "/")
			token := "token"
			_, err := getGithubRepos(context.Background(), client, Config{
				GithubUsername:     "SourceUser",
				GithubToken:        &token,
				MirrorPrivateRepos: true,
			})
			if tt.wantError {
				if err == nil || !strings.Contains(err.Error(), "GitHub token belongs to") {
					t.Fatalf("error = %v, want token owner mismatch", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("get repositories: %v", err)
			}
		})
	}
}
