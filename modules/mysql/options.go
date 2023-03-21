package mysql

type Option func(*Config)

type Config struct {
	username   string
	password   string
	database   string
	configFile string
	scripts    []string
}

func WithUsername(username string) Option {
	return func(config *Config) {
		config.username = username
	}
}

func WithPassword(password string) Option {
	return func(config *Config) {
		config.password = password
	}
}

func WithDatabase(database string) Option {
	return func(config *Config) {
		config.database = database
	}
}

func WithConfigFile(configFile string) Option {
	return func(config *Config) {
		config.configFile = configFile
	}
}

func WithScripts(scripts ...string) Option {
	return func(config *Config) {
		config.scripts = scripts
	}
}
