package couchbase

type bucket struct {
	name              string
	flushEnabled      bool
	queryPrimaryIndex bool
	quota             int
	numReplicas       int
}

func NewBucket(name string) bucket {
	return bucket{
		name:              name,
		flushEnabled:      false,
		queryPrimaryIndex: true,
		quota:             100,
		numReplicas:       0,
	}
}

func (b bucket) WithReplicas(numReplicas int) bucket {
	if numReplicas < 0 {
		numReplicas = 0
	} else if numReplicas > 3 {
		numReplicas = 3
	}

	b.numReplicas = numReplicas
	return b
}

func (b bucket) WithFlushEnabled(flushEnabled bool) bucket {
	b.flushEnabled = flushEnabled
	return b
}

func (b bucket) WithQuota(quota int) bucket {
	if quota < 100 {
		quota = 100
	}

	b.quota = quota
	return b
}

func (b bucket) WithPrimaryIndex(primaryIndex bool) bucket {
	b.queryPrimaryIndex = primaryIndex
	return b
}
