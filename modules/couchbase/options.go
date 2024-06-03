package couchbase

import "github.com/testcontainers/testcontainers-go"

// Config is the configuration for the Couchbase container, that will be stored in the container itself.
type Config struct {
	enabledServices  []Service
	username         string
	password         string
	isEnterprise     bool
	buckets          []bucket
	imageName        string
	indexStorageMode indexStorageMode
}

type credentialsCustomizer struct {
	username string
	password string
}

func (c credentialsCustomizer) Customize(req *testcontainers.Request) error {
	// NOOP, we want to simply transfer the credentials to the container
	return nil
}

// WithAdminCredentials sets the username and password for the administrator user.
func WithAdminCredentials(username, password string) credentialsCustomizer {
	return credentialsCustomizer{
		username: username,
		password: password,
	}
}

type bucketCustomizer struct {
	buckets []bucket
}

func (c bucketCustomizer) Customize(req *testcontainers.Request) error {
	// NOOP, we want to simply transfer the buckets to the container
	return nil
}

// WithBucket adds buckets to the couchbase container
func WithBuckets(bucket ...bucket) bucketCustomizer {
	return bucketCustomizer{
		buckets: bucket,
	}
}

type indexStorageCustomizer struct {
	mode indexStorageMode
}

func (c indexStorageCustomizer) Customize(req *testcontainers.Request) error {
	// NOOP, we want to simply transfer the index storage mode to the container
	return nil
}

// WithBucket adds buckets to the couchbase container
func WithIndexStorage(indexStorageMode indexStorageMode) indexStorageCustomizer {
	return indexStorageCustomizer{
		mode: indexStorageMode,
	}
}
