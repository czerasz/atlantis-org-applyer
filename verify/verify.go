package verify

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"

	"github.com/czerasz/atlantis-org-applyer/config"
	"github.com/czerasz/atlantis-org-applyer/project"
	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// Verifyer abstracts the verification process.
type Verifyer struct {
	client   *github.Client
	conf     config.Config
	projects []project.Project
	// a match table for GitHub team slug and team ID
	githubTeams              map[string]int64
	githubTeamsUserTeamCache map[string]bool
}

// Verify returns true if apply should be executed.
func (v *Verifyer) Verify(ctx context.Context) (bool, error) {
	var pr int

	prConvOnce := &sync.Once{}

	var mergeable bool

	mergeableOnce := &sync.Once{}

	for _, p := range v.projects {
		valid := p.ValidRepoOwner(v.conf.RepoOwner)
		if !valid {
			continue
		}

		valid, err := p.ValidRepoName(v.conf.RepoName)
		if err != nil {
			return false, err
		}

		if !valid {
			continue
		}

		valid, err = p.ValidProject(v.conf.AtlantisProjectName)
		if err != nil {
			return false, err
		}

		if !valid {
			continue
		}

		if p.RequiredMergeable {
			var prErr error

			prConvOnce.Do(func() {
				pr, prErr = strconv.Atoi(v.conf.PRID)
			})

			if prErr != nil {
				return false, fmt.Errorf("PR ID can not be parsed: %w", prErr)
			}

			var mergeableErr error

			mergeableOnce.Do(func() {
				mergeable, mergeableErr = isMergeable(ctx, v.client, v.conf.RepoOwner, v.conf.RepoName, pr)
			})

			if mergeableErr != nil {
				msg := fmt.Sprintf("error while checking mergable state for PR %d in %s/%s", pr, v.conf.RepoOwner, v.conf.RepoName)

				return false, fmt.Errorf("%s: %w", msg, err)
			}

			if !mergeable {
				continue
			}
		}

		allowed, err := v.applyerAllowed(ctx, p, v.conf.Username)
		if err != nil {
			return false, err
		}

		if allowed {
			return true, nil
		}
	}

	return false, nil
}

func (v *Verifyer) applyerAllowed(ctx context.Context, p project.Project, username string) (bool, error) {
	for _, u := range p.Users() {
		if u == username {
			return true, nil
		}
	}

	for _, team := range p.Teams() {
		cacheKey := fmt.Sprintf("%s/%s", team, username)
		if ok, exists := v.githubTeamsUserTeamCache[cacheKey]; exists {
			if ok {
				// only theoretical case - will not happen in real life
				// since Verify will return once applyerAllowed returns true
				return true, nil
			}

			continue
		}

		ok, err := userInGitHubTeam(ctx, v.client, v.githubTeams, username, team)
		if err != nil {
			return false, err
		}

		if ok {
			return true, nil
		}

		// cache value to not stress GitHub API too much
		v.githubTeamsUserTeamCache[cacheKey] = false
	}

	return false, nil
}

// New returns new Verifyer.
func New(ctx context.Context, c config.Config, ghc *github.Client) (*Verifyer, error) {
	p, err := loadConfig(c.ConfigPath)
	if err != nil {
		return nil, err
	}

	t, err := loadGitHubTeams(ctx, ghc, c.RepoOwner)
	if err != nil {
		return nil, err
	}

	v := Verifyer{
		client:                   ghc,
		conf:                     c,
		projects:                 p,
		githubTeams:              t,
		githubTeamsUserTeamCache: make(map[string]bool),
	}

	return &v, nil
}

func userInGitHubTeam(ctx context.Context, client *github.Client,
	teams map[string]int64, user, team string) (bool, error) {
	if teamID, ok := teams[team]; ok {
		member, resp, err := client.Teams.GetTeamMembership(ctx, teamID, user)

		// user is not in team
		if resp.StatusCode == http.StatusNotFound {
			return false, nil
		}

		if err != nil {
			return false, err
		}

		switch member.GetState() {
		case "active":
			return true, nil
		case "pending":
			// user is not yet in team
			return false, nil
		}

		return false, errors.Errorf(`unknown membership state "%s"`, member.GetState())
	}

	return false, errors.Errorf(`team "%s" not found`, team)
}

func loadConfig(fileName string) ([]project.Project, error) {
	type input struct {
		Projects []project.Project `yaml:"projects"`
	}

	i := input{}

	yamlFile, err := ioutil.ReadFile(fileName)
	if err != nil {
		return []project.Project{}, err
	}

	err = yaml.Unmarshal(yamlFile, &i)
	if err != nil {
		return []project.Project{}, err
	}

	return i.Projects, nil
}

func loadGitHubTeams(ctx context.Context, client *github.Client, org string) (map[string]int64, error) {
	teams := make(map[string]int64)
	teamsPerPage := 10

	opt := &github.ListOptions{
		PerPage: teamsPerPage,
	}

	for {
		teamList, resp, err := client.Teams.ListTeams(ctx, org, opt)
		if err != nil {
			return teams, err
		}

		for _, team := range teamList {
			teams[team.GetSlug()] = team.GetID()
		}

		if resp.NextPage == 0 {
			break
		}

		opt.Page = resp.NextPage
	}

	return teams, nil
}

func isMergeable(ctx context.Context, client *github.Client, owner, repo string, prID int) (bool, error) {
	pr, _, err := client.PullRequests.Get(ctx, owner, repo, prID)
	if err != nil {
		return false, err
	}

	// clean means - mergeable and passing commit status
	// Resource: https://docs.github.com/en/graphql/reference/enums#mergestatestatus
	return !pr.GetMerged() && pr.GetMergeable() && pr.GetMergeableState() == "clean", nil
}
