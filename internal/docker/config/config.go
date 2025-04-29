package config

const (
	// EnvOverrideDir is the name of the environment variable that can be
	// used to override the location of the client configuration files (~/.docker).
	//
	// It takes priority over the default.
	EnvOverrideDir = "DOCKER_CONFIG"

	// configFileDir is the name of the directory containing the client configuration files
	configFileDir = ".docker"

	// configFileName is the name of the client configuration file inside the
	// config-directory.
	FileName = "config.json"
)
