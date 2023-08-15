package mkdocs

func UpdateConfig(configFile string, isModule bool, exampleMd string, indexMd string) error {
	config, err := ReadConfig(configFile)
	if err != nil {
		return err
	}
	config.addExample(isModule, exampleMd, indexMd)
	return writeConfig(configFile, config)
}

func CopyConfig(configFile string, tmpFile string) error {
	config, err := ReadConfig(configFile)
	if err != nil {
		return err
	}
	return writeConfig(tmpFile, config)
}
