package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v63/github"
)

type ListOptions struct {
	Affiliation string
	PerPage     int
	Page        int
}

func RepositoryList(ctx context.Context, client *github.Client, user string, opt *ListOptions,
) ([]*github.Repository, *github.Response, error) {
	if opt == nil {
		opt = &ListOptions{}
	}
	if user == "" {
		_opts := &github.RepositoryListByAuthenticatedUserOptions{
			Affiliation: opt.Affiliation,
			ListOptions: github.ListOptions{
				PerPage: opt.PerPage,
				Page:    opt.Page,
			},
		}
		return client.Repositories.ListByAuthenticatedUser(ctx, _opts)
	}
	_opts := &github.RepositoryListByUserOptions{
		ListOptions: github.ListOptions{
			PerPage: opt.PerPage,
			Page:    opt.Page,
		},
	}
	return client.Repositories.ListByUser(ctx, user, _opts)
}

func GetGithubRepos(ctx context.Context, config Config) ([]*github.Repository, error) {
	client := github.NewClient(nil)
	return getGithubRepos(ctx, client, config)
}

func getGithubRepos(ctx context.Context, client *github.Client, config Config) ([]*github.Repository, error) {
	var repos []*github.Repository
	if config.GithubToken != nil {
		client = client.WithAuthToken(*config.GithubToken)
	}

	opt := &ListOptions{
		PerPage: 30,
	}

	username := config.GithubUsername
	if config.MirrorPrivateRepos {
		authenticatedUser, _, err := client.Users.Get(ctx, "")
		if err != nil {
			return nil, fmt.Errorf("get authenticated GitHub user: %w", err)
		}
		if !strings.EqualFold(authenticatedUser.GetLogin(), config.GithubUsername) {
			return nil, fmt.Errorf("GitHub token belongs to %q, expected %q", authenticatedUser.GetLogin(), config.GithubUsername)
		}
		username = ""
		opt.Affiliation = "owner"
	}

	for {
		pageRepos, resp, err := RepositoryList(
			ctx, client, username, opt)
		if err != nil {
			return nil, err
		}
		repos = append(repos, pageRepos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	// Filter out repos that we are not mirroring
	if !config.MirrorForks || !config.MirrorPublicRepos || !config.MirrorPrivateRepos {
		var filteredRepos []*github.Repository
		for _, repo := range repos {
			if !config.MirrorForks && *repo.Fork {
				continue
			}
			if !config.MirrorPublicRepos && (!*repo.Private && !*repo.Fork) {
				continue
			}
			if !config.MirrorPrivateRepos && *repo.Private {
				continue
			}
			filteredRepos = append(filteredRepos, repo)
		}
		repos = filteredRepos
	}

	return repos, nil
}
