package config

import (
	"github.com/kelseyhightower/envconfig"
)

// Config represents the configuration object
type Config struct {
	// environment variables provided by Atlantis
	// resources:
	// - documentation: https://www.runatlantis.io/docs/custom-workflows.html#reference
	// - code: https://github.com/runatlantis/atlantis/blob/2ac77dbe5873cce909d847886ac269c28804bb46/server/events/runtime/run_step_runner.go#L40-L56
	GitHubBaseURL string `envconfig:"GITHUB_BASE_URL"`
	GitHubToken   string `required:"true" envconfig:"GITHUB_TOKEN"`

	RepoOwner string `required:"true" envconfig:"BASE_REPO_OWNER"`
	RepoName  string `required:"true" envconfig:"BASE_REPO_NAME"`
	PRID      string `envconfig:"PULL_NUM"`

	AtlantisProjectName string `required:"true" envconfig:"PROJECT_NAME"`
	Username            string `required:"true" envconfig:"USER_NAME"`

	ConfigPath string `default:"atlantis-org-applyer.yaml" envconfig:"CONFIG_PATH"`
}

// New returns configuration
func New() (Config, error) {
	var c Config
	err := envconfig.Process("atlantis-org-applyer", &c)

	return c, err
}
