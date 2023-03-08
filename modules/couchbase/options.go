package couchbase

type Option func(*Config)

type Config struct {
	enabledServices  []service
	username         string
	password         string
	isEnterprise     bool
	buckets          []bucket
	imageName        string
	indexStorageMode indexStorageMode
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

func WithBucket(bucket bucket) Option {
	return func(c *Config) {
		c.buckets = append(c.buckets, bucket)
	}
}

func WithImageName(imageName string) Option {
	return func(c *Config) {
		c.imageName = imageName
	}
}

func WithIndexStorageMode(indexStorageMode indexStorageMode) Option {
	return func(c *Config) {
		c.indexStorageMode = indexStorageMode
	}
}
