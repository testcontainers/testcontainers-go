package couchbase

// Option is a function that configures the Couchbase container.
type Option func(*Config)

// Config is the configuration for the Couchbase container, that will be stored in the container itself.
type Config struct {
	enabledServices  []service
	username         string
	password         string
	isEnterprise     bool
	buckets          []bucket
	imageName        string
	indexStorageMode indexStorageMode
}

// WithEnterpriseService enables the eventing service in the container.
// Only available in the Enterprise Edition of Couchbase Server.
func WithEventingService() Option {
	return func(c *Config) {
		c.enabledServices = append(c.enabledServices, eventing)
	}
}

// WithAnalyticsService enables the analytics service in the container.
// Only available in the Enterprise Edition of Couchbase Server.
func WithAnalyticsService() Option {
	return func(c *Config) {
		c.enabledServices = append(c.enabledServices, analytics)
	}
}

// WithCredentials sets the username and password for the administrator user.
func WithCredentials(username, password string) Option {
	return func(c *Config) {
		c.username = username
		c.password = password
	}
}

// WithBucket adds a bucket to the container.
func WithBucket(bucket bucket) Option {
	return func(c *Config) {
		c.buckets = append(c.buckets, bucket)
	}
}

// WithImageName allows to override the default image name.
func WithImageName(imageName string) Option {
	return func(c *Config) {
		c.imageName = imageName
	}
}

// WithIndexStorageMode sets the storage mode to be used in the cluster.
func WithIndexStorageMode(indexStorageMode indexStorageMode) Option {
	return func(c *Config) {
		c.indexStorageMode = indexStorageMode
	}
}
