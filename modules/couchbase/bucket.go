package couchbase

type bucket struct {
	name              string
	flushEnabled      bool
	queryPrimaryIndex bool
	quota             int
	numReplicas       int
}

// NewBucket creates a new bucket with the given name, using default values for all other fields.
func NewBucket(name string) bucket {
	return bucket{
		name:              name,
		flushEnabled:      false,
		queryPrimaryIndex: true,
		quota:             100,
		numReplicas:       0,
	}
}

// WithReplicas sets the number of replicas for this bucket. The minimum value is 0 and the maximum value is 3.
func (b bucket) WithReplicas(numReplicas int) bucket {
	if numReplicas < 0 {
		numReplicas = 0
	} else if numReplicas > 3 {
		numReplicas = 3
	}

	b.numReplicas = numReplicas
	return b
}

// WithFlushEnabled sets whether the bucket should be flushed when the container is stopped.
func (b bucket) WithFlushEnabled(flushEnabled bool) bucket {
	b.flushEnabled = flushEnabled
	return b
}

// WithQuota sets the bucket quota in megabytes. The minimum value is 100 MB.
func (b bucket) WithQuota(quota int) bucket {
	if quota < 100 { // Couchbase Server 6.5.0 has a minimum of 100 MB
		quota = 100
	}

	b.quota = quota
	return b
}

// WithPrimaryIndex sets whether the primary index should be created for this bucket.
func (b bucket) WithPrimaryIndex(primaryIndex bool) bucket {
	b.queryPrimaryIndex = primaryIndex
	return b
}
