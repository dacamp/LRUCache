package cache

import (
	"time"
)

// A Cache is a generalized interface to a cache.  See LRU for a specific
// implementation (bounded cache with LRU eviction)
type Cache interface {
	// Get retrieves an element based on a key, returning nil if the element
	// does not exist
	Get(key string) interface{}

	// Put adds an element to the cache, returning the previous element
	Put(key string, value interface{}) interface{}

	// Delete deletes an element in the cache
	Delete(key string)
}

// Options control the behavior of the cache
type Options struct {
	// TTL controls the time-to-live for a given cache entry.  Cache entries that
	// are older than the TTL will not be returned
	TTL time.Duration

	// InitialCapacity controls the initial capacity of the cache
	InitialCapacity int

	// RemovedFunc is an optional function called when an element
	// is scheduled for deletion
	RemovedFunc RemovedFunc
}

// RemovedFunc is a type for notifying applications when an item is
// scheduled for removal from the Cache. If f is a function with the
// appropriate signature and i is the interface{} scheduled for
// deletion, Cache calls go f(i)
type RemovedFunc func(interface{})
