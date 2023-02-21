package couchbase

type Option func(*Config)

type Config struct {
	enabledServices []service
	username        string
	password        string
	isEnterprise    bool
}

func WithEventingService() Option {
	return func(c *Config) {
		c.enabledServices = append(c.enabledServices, eventing)
	}
}

func WithAnalyticsService() Option {
	return func(c *Config) {
		c.enabledServices = append(c.enabledServices, analytics)
	}
}

func WithCredentials(username, password string) Option {
	return func(c *Config) {
		c.username = username
		c.password = password
	}
}
