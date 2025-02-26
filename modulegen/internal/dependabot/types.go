package dependabot

import (
	"sort"
)

// Config the configuration for the dependabot config file
type Config struct {
	Version int     `yaml:"version"`
	Updates Updates `yaml:"updates"`
}

// Schedule the schedule for the dependabot config file
type Schedule struct {
	Interval string `yaml:"interval"`
	Day      string `yaml:"day"`
}

// Update the update for the dependabot config file
type Update struct {
	PackageEcosystem      string   `yaml:"package-ecosystem"`
	Directory             string   `yaml:"directory"`
	Schedule              Schedule `yaml:"schedule"`
	OpenPullRequestsLimit int      `yaml:"open-pull-requests-limit"`
	RebaseStrategy        string   `yaml:"rebase-strategy"`
}

// Updates the updates for the dependabot config file
type Updates []Update

func newUpdate(directory string, packageExosystem string) Update {
	return Update{
		Directory:             directory,
		OpenPullRequestsLimit: 3,
		PackageEcosystem:      packageExosystem,
		RebaseStrategy:        "disabled",
		Schedule: Schedule{
			Interval: "monthly",
			Day:      "sunday",
		},
	}
}

// addUpdate adds an update to the config
func (c *Config) addUpdate(newUpdate Update) {
	exists := false
	for _, update := range c.Updates {
		if update.Directory == newUpdate.Directory && update.PackageEcosystem == newUpdate.PackageEcosystem {
			exists = true
			break
		}
	}

	if !exists {
		c.Updates = append(c.Updates, newUpdate)
		sort.Sort(c.Updates)
	}
}
