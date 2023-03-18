package mysql

type Option func(*Config)

type Config struct {
	username   string
	password   string
	database   string
	configFile string
}

func WithUsername(username string) Option {
	return func(config *Config) {
		config.username = username
	}
}

func withPassword(password string) Option {
	return func(config *Config) {
		config.password = password
	}
}

func withDatabase(database string) Option {
	return func(config *Config) {
		config.database = database
	}
}

func withConfigFile(configFile string) Option {
	return func(config *Config) {
		config.configFile = configFile
	}
}
