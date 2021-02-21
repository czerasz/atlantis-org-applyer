package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/czerasz/atlantis-org-applyer/config"
	"github.com/czerasz/atlantis-org-applyer/verify"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

func main() {
	c, err := config.New()
	if err != nil {
		log.Fatalf("error while loading the config: %s", err)
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: c.GitHubToken,
		},
	)
	tc := oauth2.NewClient(ctx, ts)

	var ghc *github.Client
	if strings.TrimSpace(c.GitHubHostname) == "" {
		ghc = github.NewClient(tc)
	} else {
		baseURL := fmt.Sprintf("https://%s", c.GitHubHostname)
		ghc, err = github.NewEnterpriseClient(baseURL, baseURL, tc)
		if err != nil {
			log.Fatalf("can't create client: %v", err)
		}
	}

	v, err := verify.New(ctx, c, ghc)
	if err != nil {
		log.Fatalf("error while creating verifyer: %s", err)
	}

	allowed, err := v.Verify(ctx)
	if err != nil {
		log.Fatalf("error during verification: %s", err)
	}

	if !allowed {
		log.Printf(`user "%s" can not apply Atlantis project "%s" on PR#%s for repository %s/%s`, c.Username, c.AtlantisProjectName, c.PRID, c.RepoOwner, c.RepoName)
		os.Exit(1)
	}
}
