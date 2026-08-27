package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
)

type Config struct {
	GithubUsername     string
	ForgejoUrl         string
	ForgejoToken       string
	ForgejoUsername    string
	GithubToken        *string
	NumWorkers         int
	MirrorPublicRepos  bool
	MirrorPrivateRepos bool
	MirrorForks        bool
	DryRun             bool
	SyncInterval       int
	WebAddr            string
}

func (c *Config) log() {
	log.Printf("Configuration:\n")
	log.Printf("  Github Username: %s\n", c.GithubUsername)
	log.Printf("  Forgejo URL: %s\n", c.ForgejoUrl)
	log.Printf("  Forgejo Token: ****\n")
	log.Printf("  Forgejo Username: %s\n", c.ForgejoUsername)
	if c.GithubToken != nil {
		log.Printf("  Github Token: ****\n") // Dereference pointer to print value
	} else {
		log.Printf("  Github Token: Not provided\n")
	}
	log.Printf("  Mirror Public Repos: %t\n", c.MirrorPublicRepos)
	log.Printf("  Mirror Private Repos: %t\n", c.MirrorPrivateRepos)
	log.Printf("  Mirror Forks: %t\n", c.MirrorForks)
	log.Printf("  SyncInterval: %d (seconds)\n", c.SyncInterval)
	log.Printf("  Web Address: %s\n", c.WebAddr)
	log.Printf("  Number of Workers: %d\n", c.NumWorkers)
	log.Printf("  Dry Run: %t\n", c.DryRun)
}

func (c *Config) resolve() error {
	if c.ForgejoUsername == "" {
		client, err := ForgejoClient(context.Background(), *c)
		if err != nil {
			return fmt.Errorf("error creating Forgejo client: %w\n", err)
		}
		username, err := ForgejoGetUsername(client)
		if err != nil {
			return fmt.Errorf("error getting Forgejo username: %w", err)
		}
		c.ForgejoUsername = username
	}
	return nil
}

func (c *Config) validate() error {
	errors := []string{}
	if c.GithubUsername == "" {
		errors = append(errors, "GITHUB_USERNAME environment variable not set")
	}
	if c.ForgejoUrl == "" {
		errors = append(errors, "FORGEJO_URL environment variable not set")
	}
	if c.ForgejoToken == "" {
		errors = append(errors, "FORGEJO_TOKEN environment variable not set")
	}
	if c.MirrorPrivateRepos && c.GithubToken == nil {
		errors = append(errors, "GITHUB_TOKEN environment variable not set (required for mirroring private repos)")
	}
	if len(errors) > 0 {
		return fmt.Errorf("config validation errors: %s", strings.Join(errors, ", "))
	}
	return nil
}

func MakeConfigFromEnv() (Config, error) {
	var envErrors []error

	githubUsername, err := GetEnvStrict("GITHUB_USERNAME")
	if err != nil {
		envErrors = append(envErrors, err)
	}
	forgejoUrl, err := GetEnvStrict("FORGEJO_URL")
	if err != nil {
		envErrors = append(envErrors, err)
	}
	forgejoToken, err := GetEnvStrict("FORGEJO_TOKEN")
	if err != nil {
		envErrors = append(envErrors, err)
	}
	if len(envErrors) > 0 {
		return Config{}, errors.Join(envErrors...)
	}

	c := Config{
		GithubUsername:     githubUsername,
		ForgejoUrl:         forgejoUrl,
		ForgejoToken:       forgejoToken,
		GithubToken:        GetEnvOptional("GITHUB_TOKEN"),
		NumWorkers:         GetEnvInt("HUBTOJO_NUM_WORKERS", 5),
		MirrorPublicRepos:  GetEnvBool("HUBTOJO_MIRROR_PUBLIC_REPOS", true),
		MirrorPrivateRepos: GetEnvBool("HUBTOJO_MIRROR_PRIVATE_REPOS", false),
		MirrorForks:        GetEnvBool("HUBTOJO_MIRROR_FORKS", false),
		DryRun:             GetEnvBool("HUBTOJO_DRY_RUN", false),
		SyncInterval:       GetEnvInt("HUBTOJO_SYNC_INTERVAL", 3600),
		WebAddr:            GetEnvString("HUBTOJO_WEB_ADDR", ":8080"),
	}
	err = c.validate()
	if err != nil {
		return Config{}, fmt.Errorf("error validating config: %w", err)
	}
	err = c.resolve()
	if err != nil {
		return Config{}, fmt.Errorf("error resolving config: %w", err)
	}
	return c, nil
}
