package couchbase

type indexStorageMode string

const (
	MemoryOptimized indexStorageMode = "memory_optimized"
	Plasma          indexStorageMode = "plasma"
	ForestDB        indexStorageMode = "forestdb"
)
