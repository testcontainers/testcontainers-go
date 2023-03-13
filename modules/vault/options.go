package vault

// Option is a function type for modifying a Vault configuration
type Option func(*Config)

// Config is a struct that contains various configuration options for the Vault
type Config struct {
	imageName    string              // The name of the Docker image to use for the Vault
	token        string              // The root token to use for the Vault
	port         int                 // The port number to use for the Vault
	secrets      map[string][]string // A map of secret paths to their respective secret values
	initCommands []string            // A list of commands to execute on Vault initialization
	logLevel     LogLevel            // The level of logging to use for the Vault
}

// WithImageName is an option function that sets the Docker image name for the Vault
func WithImageName(imageName string) Option {
	return func(c *Config) {
		c.imageName = imageName
	}
}

// WithToken is an option function that sets the root token for the Vault
func WithToken(token string) Option {
	return func(c *Config) {
		c.token = token
	}
}

// WithPort is an option function that sets the port number for the Vault
func WithPort(port int) Option {
	return func(c *Config) {
		c.port = port
	}
}

// WithSecrets is an option function that adds a set of secrets to the Vault's configuration
func WithSecrets(path, firstSecret string, remainingSecrets ...string) Option {
	return func(c *Config) {
		secretList := []string{firstSecret}

		for _, secret := range remainingSecrets {
			secretList = append(secretList, secret)
		}

		if secrets, ok := c.secrets[path]; ok {
			secretList = append(secretList, secrets...)
		}

		c.secrets[path] = secretList
	}
}

// WithInitCommands is an option function that adds a set of initialization commands to the Vault's configuration
func WithInitCommands(commands ...string) Option {
	return func(c *Config) {
		c.initCommands = append(c.initCommands, commands...)
	}
}

// WithLogLevel is an option function that sets the logging level for the Vault
func WithLogLevel(logLevel LogLevel) Option {
	return func(c *Config) {
		c.logLevel = logLevel
	}
}
