package main

import (
	"context"
	"os"
	"strings"

	"github.com/czerasz/atlantis-org-applyer/config"
	"github.com/czerasz/atlantis-org-applyer/verify"
	"github.com/google/go-github/github"
	logrus "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

func main() {
	log := initLogger()

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
	if strings.TrimSpace(c.GitHubBaseURL) == "" {
		ghc = github.NewClient(tc)
	} else {
		ghc, err = github.NewEnterpriseClient(c.GitHubBaseURL, c.GitHubBaseURL, tc)
		if err != nil {
			log.Fatalf("can't create client: %v", err)
		}
	}

	v, err := verify.New(ctx, c, ghc)
	if err != nil {
		log.Fatalf("error while creating verifyer: %s", err)
	}

	allowed, err := v.Verify(ctx, log)
	if err != nil {
		log.Fatalf("error during verification: %s", err)
	}

	if !allowed {
		msg := `user "%s" can not apply Atlantis project "%s" on PR#%s for repository %s/%s`
		log.Printf(msg, c.Username, c.AtlantisProjectName, c.PRID, c.RepoOwner, c.RepoName)
		os.Exit(1)
	}
}

// initLogger initializes the logger with "debug" level by default.
func initLogger() *logrus.Logger {
	lvl, ok := os.LookupEnv("LOG_LEVEL")

	if !ok {
		lvl = "debug"
	}

	ll, err := logrus.ParseLevel(lvl)
	if err != nil {
		ll = logrus.DebugLevel
	}

	log := logrus.New()
	log.SetLevel(ll)

	return log
}
