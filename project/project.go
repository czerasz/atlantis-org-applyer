package project

import "regexp"

// Applyer represents the applyer structure
type Applyer struct {
	Name string `yaml:"name"`
	Type string `yaml:"type,omitempty"`
}

// Project represents the Atlantis project item
type Project struct {
	RepoOwner         string    `yaml:"repo_owner"`
	RepoName          string    `yaml:"repo_name"`
	Project           string    `yaml:"atlantis_project"`
	RequiredMergeable bool      `yaml:"required_mergeable"`
	Applyers          []Applyer `yaml:"applyers"`
}

// ValidRepoOwner returns true if repository owner is valid
func (p *Project) ValidRepoOwner(repoOwner string) bool {
	return p.RepoOwner == repoOwner
}

// ValidRepoName returns true if repository name is valid
func (p *Project) ValidRepoName(repoName string) (bool, error) {
	r, err := regexp.Compile(p.RepoName)
	if err != nil {
		return false, err
	}

	return r.Match([]byte(repoName)), nil
}

// ValidProject returns true if projects regular expression is valid
func (p *Project) ValidProject(project string) (bool, error) {
	r, err := regexp.Compile(p.Project)
	if err != nil {
		return false, err
	}

	return r.Match([]byte(project)), nil
}

// Teams filters applyers for teams
func (p *Project) Teams() (teams []string) {
	teams = []string{}
	for _, i := range p.Applyers {
		if i.Type == "team" {
			teams = append(teams, i.Name)
		}
	}

	return
}

// Users filters applyers for users
func (p *Project) Users() (users []string) {
	users = []string{}

	for _, i := range p.Applyers {
		if i.Type == "" || i.Type == "user" {
			users = append(users, i.Name)
		}
	}

	return
}
