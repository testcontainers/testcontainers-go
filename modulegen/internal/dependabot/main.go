package dependabot

import (
	"fmt"

	"github.com/testcontainers/testcontainers-go/modulegen/internal/context"
)

// Generator is a struct that contains the logic to generate the dependabot config file.
type Generator struct{}

// AddModule update dependabot with the new module
func (g Generator) AddModule(ctx context.Context, tcModule context.TestcontainersModule) error {
	configFile := ctx.DependabotConfigFile()

	config, err := readConfig(configFile)
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	packageEcosystem := "gomod"
	directory := "/" + tcModule.ParentDir() + "/" + tcModule.Lower()

	config.addUpdate(newUpdate(directory, packageEcosystem))

	return writeConfig(configFile, config)
}

// Generate generates dependabot config file from source
func (g Generator) Generate(ctx context.Context, examples []string, modules []string) error {
	configFile := ctx.DependabotConfigFile()

	config, err := readConfig(configFile)
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	for _, example := range examples {
		directory := "/examples/" + example
		config.addUpdate(newUpdate(directory, "gomod"))
	}

	for _, module := range modules {
		directory := "/modules/" + module
		config.addUpdate(newUpdate(directory, "gomod"))
	}

	return writeConfig(configFile, config)
}

// GetUpdates returns the updates from the dependabot config file
func GetUpdates(configFile string) (Updates, error) {
	config, err := readConfig(configFile)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	return config.Updates, nil
}

// CopyConfig helper function to copy the dependabot config file to a another file
// in the tests.
func CopyConfig(configFile string, tmpFile string) error {
	config, err := readConfig(configFile)
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}
	return writeConfig(tmpFile, config)
}
