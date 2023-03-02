package main

import (
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const updateSchedule = "monthly"

type Updates []Update

type DependabotConfig struct {
	Version int     `yaml:"version"`
	Updates Updates `yaml:"updates"`
}

type Schedule struct {
	Interval string `yaml:"interval"`
}

type Update struct {
	PackageEcosystem      string   `yaml:"package-ecosystem"`
	Directory             string   `yaml:"directory"`
	Schedule              Schedule `yaml:"schedule"`
	OpenPullRequestsLimit int      `yaml:"open-pull-requests-limit"`
	RebaseStrategy        string   `yaml:"rebase-strategy"`
}

func NewUpdate(example Example) Update {
	return Update{
		Directory:             "/" + example.ParentDir() + "/" + example.Lower(),
		OpenPullRequestsLimit: 3,
		PackageEcosystem:      "gomod",
		RebaseStrategy:        "disabled",
		Schedule: Schedule{
			Interval: updateSchedule,
		},
	}
}

// Len is the number of elements in the collection.
func (u Updates) Len() int {
	return len(u)
}

// Less reports whether the element with index i
// must sort before the element with index j.
//
// If both Less(i, j) and Less(j, i) are false,
// then the elements at index i and j are considered equal.
// Sort may place equal elements in any order in the final result,
// while Stable preserves the original input order of equal elements.
//
// Less must describe a transitive ordering:
//  - if both Less(i, j) and Less(j, k) are true, then Less(i, k) must be true as well.
//  - if both Less(i, j) and Less(j, k) are false, then Less(i, k) must be false as well.
//
// Note that floating-point comparison (the < operator on float32 or float64 values)
// is not a transitive ordering when not-a-number (NaN) values are involved.
// See Float64Slice.Less for a correct implementation for floating-point values.
func (u Updates) Less(i, j int) bool {
	return u[i].Directory < u[j].Directory
}

// Swap swaps the elements with indexes i and j.
func (u Updates) Swap(i, j int) {
	u[i], u[j] = u[j], u[i]
}

func getDependabotConfigFile(rootDir string) string {
	return filepath.Join(rootDir, ".github", "dependabot.yml")
}

func getDependabotUpdates() ([]Update, error) {
	parent, err := getRootDir()
	if err != nil {
		return nil, err
	}

	config, err := readDependabotConfig(parent)
	if err != nil {
		return nil, err
	}

	return config.Updates, nil
}

func readDependabotConfig(rootDir string) (*DependabotConfig, error) {
	configFile := getDependabotConfigFile(rootDir)

	file, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	config := &DependabotConfig{}

	err = yaml.Unmarshal(file, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func writeDependabotConfig(rootDir string, config *DependabotConfig) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	file := getDependabotConfigFile(rootDir)

	return ioutil.WriteFile(file, data, 0777)
}
