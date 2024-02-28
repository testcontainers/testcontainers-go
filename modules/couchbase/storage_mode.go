package couchbase

// The storage mode to be used for all global secondary indexes in the cluster.
// Please note: "plasma" and "memory optimized" are options in the Enterprise Edition of Couchbase Server. If you are
// using the Community Edition, the only value allowed is forestdb.
type indexStorageMode string

// storageTypes {

const (
	// MemoryOptimized sets the cluster-wide index storage mode to use memory optimized global
	// secondary indexes which can perform index maintenance and index scan faster at in-memory speeds.
	// This is the default value for the testcontainers couchbase implementation.
	MemoryOptimized indexStorageMode = "memory_optimized"

	// Plasma sets the cluster-wide index storage mode to use the Plasma storage engine,
	// which can utilize both memory and persistent storage for index maintenance and index scans.
	Plasma indexStorageMode = "plasma"

	// ForestDB sets the cluster-wide index storage mode to use the forestdb storage engine,
	// which only utilizes persistent storage for index maintenance and scans. It is the only option available
	// for the community edition.
	ForestDB indexStorageMode = "forestdb"
)

// }
