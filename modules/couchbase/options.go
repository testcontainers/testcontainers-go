package couchbase

type Option func(*Config)

type Config struct {
	enabledServices []service
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
