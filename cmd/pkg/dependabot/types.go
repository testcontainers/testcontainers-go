package dependabot

import (
	"sort"
)

type Updates []Update

type Config struct {
	Version int     `yaml:"version"`
	Updates Updates `yaml:"updates"`
}

type Group struct {
	Patterns []string `yaml:"patterns"`
}

type Groups map[string]Group

type Schedule struct {
	Interval string `yaml:"interval"`
	Day      string `yaml:"day"`
}

type Update struct {
	PackageEcosystem      string   `yaml:"package-ecosystem"`
	Directory             string   `yaml:"directory"`
	Schedule              Schedule `yaml:"schedule"`
	OpenPullRequestsLimit int      `yaml:"open-pull-requests-limit"`
	RebaseStrategy        string   `yaml:"rebase-strategy"`
	Groups                Groups   `yaml:"groups,omitempty"`
}

func NewUpdate(directory string, packageExosystem string) Update {
	return Update{
		Directory:             directory,
		OpenPullRequestsLimit: 3,
		PackageEcosystem:      packageExosystem,
		RebaseStrategy:        "disabled",
		Schedule: Schedule{
			Interval: "monthly",
			Day:      "sunday",
		},
		Groups: Groups{
			"all": Group{
				Patterns: []string{"*"},
			},
		},
	}
}

func (c *Config) AddExampleFromContext(ctx *Context) {
	exists := false
	newUpdate := NewUpdate(ctx.Directory(), "gomod")
	for _, update := range c.Updates {
		if update.Directory == newUpdate.Directory && update.PackageEcosystem == newUpdate.PackageEcosystem {
			exists = true
		}
	}

	if !exists {
		c.Updates = append(c.Updates, newUpdate)
		sort.Sort(c.Updates)
	}
}
