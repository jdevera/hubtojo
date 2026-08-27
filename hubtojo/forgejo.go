package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v3"
	"github.com/google/go-github/v63/github"
)

type MirrorResult int

const (
	Created MirrorResult = iota
	WouldCreate
	Skipped
	Failed
)

func ForgejoGetUsername(client *forgejo.Client) (string, error) {
	user, _, err := client.GetMyUserInfo()
	if err != nil {
		return "", err
	}
	return user.UserName, nil
}

func ForgejoClient(ctx context.Context, config Config) (*forgejo.Client, error) {
	return forgejo.NewClient(config.ForgejoUrl, forgejo.SetToken(config.ForgejoToken), forgejo.SetContext(ctx))
}

// ForgejoMirror creates a repository on Forgejo for the given GitHub repository. The
// repository is created with the same name and description as the GitHub
// repository. The repository is created as a mirror of the GitHub repository.
func ForgejoMirror(ctx context.Context, githubRepo *github.Repository, config Config) (MirrorResult, error) {
	prefix := ""
	if workerId := ctx.Value("worker_id"); workerId != nil {
		prefix = fmt.Sprintf("[Worker %d] ", workerId)
	}
	client, err := forgejo.NewClient(config.ForgejoUrl,
		forgejo.SetToken(config.ForgejoToken),
		forgejo.SetContext(ctx),
	)
	if err != nil {
		return Failed, err
	}
	forgejoRepo, resp, err := client.GetRepo(config.ForgejoUsername, *githubRepo.Name)
	if err == nil {
		log.Printf("%sSkipping repository %s. It already exists on Forgejo\n", prefix, forgejoRepo.FullName)
		return Skipped, nil
	}
	if resp == nil || resp.StatusCode != http.StatusNotFound {
		return Failed, fmt.Errorf("check Forgejo repository %s/%s: %w", config.ForgejoUsername, *githubRepo.Name, err)
	}
	if config.DryRun {
		log.Printf("%s[DRY-RUN] Would create repository %s on Forgejo\n", prefix, *githubRepo.FullName)
		return WouldCreate, nil
	}
	log.Printf("%sCreating repository %s on Forgejo\n", prefix, *githubRepo.FullName)

	githubAuth := ""
	if config.GithubToken != nil {
		githubAuth = *config.GithubToken
	}
	option := forgejo.MigrateRepoOption{
		AuthToken: githubAuth,
		CloneAddr: *githubRepo.CloneURL,
		RepoName:  *githubRepo.Name,
		RepoOwner: config.ForgejoUsername,
		Private:   *githubRepo.Private,
		Mirror:    true,
	}
	if githubRepo.Description != nil {
		option.Description = *githubRepo.Description
	}
	_, _, err = client.MigrateRepo(option)
	if err != nil {
		return Failed, err
	}
	log.Printf("%sRepository %s created on Forgejo\n", prefix, *githubRepo.FullName)
	return Created, nil
}
