package cache

import (
	"container/list"
	"sync"
	"time"
)

// lru is a concurrent fixed size cache that evicts elements in lru order
type lru struct {
	mut      sync.Mutex
	byAccess *list.List
	byKey    map[string]*list.Element
	maxSize  int
	ttl      time.Duration
	rmFunc   RemovedFunc
}

type cacheEntry struct {
	key        string
	expiration time.Time
	value      interface{}
}

// New creates a new cache with the given options
func New(maxSize int, opts *Options) Cache {
	if opts == nil {
		opts = &Options{}
	}

	return &lru{
		byAccess: list.New(),
		byKey:    make(map[string]*list.Element, opts.InitialCapacity),
		ttl:      opts.TTL,
		maxSize:  maxSize,
		rmFunc:   opts.RemovedFunc,
	}
}

// NewLRU creates a new LRU cache of the given size, setting initial capacity
// to the max size
func NewLRU(maxSize int) Cache {
	return New(maxSize, nil)
}

// NewLRUWithInitialCapacity creates a new LRU cache with an initial capacity
// and a max size
func NewLRUWithInitialCapacity(initialCapacity, maxSize int) Cache {
	return New(maxSize, &Options{
		InitialCapacity: initialCapacity,
	})
}

// Get retrieves the value stored under the given key
func (c *lru) Get(key string) interface{} {
	c.mut.Lock()
	defer c.mut.Unlock()

	elt := c.byKey[key]
	if elt == nil {
		return nil
	}

	cacheEntry := elt.Value.(*cacheEntry)
	if !cacheEntry.expiration.IsZero() && time.Now().After(cacheEntry.expiration) {
		// Entry has expired
		if c.rmFunc != nil {
			go c.rmFunc(cacheEntry.value)
		}
		c.byAccess.Remove(elt)
		delete(c.byKey, cacheEntry.key)
		return nil
	}

	c.byAccess.MoveToFront(elt)
	return cacheEntry.value
}

// Put puts a new value associated with a given key, returning the existing value (if present)
func (c *lru) Put(key string, value interface{}) interface{} {
	c.mut.Lock()
	defer c.mut.Unlock()

	elt := c.byKey[key]
	if elt != nil {
		entry := elt.Value.(*cacheEntry)
		existing := entry.value
		entry.value = value
		if c.ttl != 0 {
			entry.expiration = time.Now().Add(c.ttl)
		}
		c.byAccess.MoveToFront(elt)
		return existing
	}

	entry := &cacheEntry{
		key:   key,
		value: value,
	}

	if c.ttl != 0 {
		entry.expiration = time.Now().Add(c.ttl)
	}
	c.byKey[key] = c.byAccess.PushFront(entry)
	if len(c.byKey) == c.maxSize {
		oldest := c.byAccess.Remove(c.byAccess.Back()).(*cacheEntry)
		if c.rmFunc != nil {
			go c.rmFunc(oldest.value)
		}
		delete(c.byKey, oldest.key)
	}

	return nil
}

// Delete deletes a key, value pair associated with a key
func (c *lru) Delete(key string) {
	c.mut.Lock()
	defer c.mut.Unlock()

	elt := c.byKey[key]
	if elt != nil {
		entry := c.byAccess.Remove(elt).(*cacheEntry)
		if c.rmFunc != nil {
			go c.rmFunc(entry.value)
		}
		delete(c.byKey, key)
	}
}
