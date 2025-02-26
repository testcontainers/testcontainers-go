package dependabot

import (
	"fmt"

	"github.com/testcontainers/testcontainers-go/modulegen/internal/context"
)

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

// Refresh refresh the dependabot config file for all the modules
func (g Generator) Refresh(ctx context.Context, tcModules []context.TestcontainersModule) error {
	configFile := ctx.DependabotConfigFile()

	config, err := readConfig(configFile)
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	for _, tcModule := range tcModules {
		directory := "/" + tcModule.ParentDir() + "/" + tcModule.Lower()
		config.addUpdate(newUpdate(directory, "gomod"))
	}

	return writeConfig(configFile, config)
}

// Generate generates dependabot config file from source
func (g Generator) Generate(ctx context.Context) error {
	configFile := ctx.DependabotConfigFile()

	config, err := readConfig(configFile)
	if err != nil {
		return err
	}

	return writeConfig(configFile, config)
}

func GetUpdates(configFile string) (Updates, error) {
	config, err := readConfig(configFile)
	if err != nil {
		return nil, err
	}
	return config.Updates, nil
}

func CopyConfig(configFile string, tmpFile string) error {
	config, err := readConfig(configFile)
	if err != nil {
		return err
	}
	return writeConfig(tmpFile, config)
}
