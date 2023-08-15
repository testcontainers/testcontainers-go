package dependabot

func UpdateConfig(configFile string, directory string, packageEcosystem string) error {
	config, err := readConfig(configFile)
	if err != nil {
		return err
	}
	config.addUpdate(newUpdate(directory, packageEcosystem))
	return writeConfig(configFile, config)
}

func GetUpdates(configFile string) (Updates, error) {
	config, err := readConfig(configFile)
	if err != nil {
		return nil, err
	}
	return config.Updates, nil
}

func CopyDependabotConfig(configFile string, tmpFile string) error {
	config, err := readConfig(configFile)
	if err != nil {
		return err
	}
	return writeConfig(tmpFile, config)
}
