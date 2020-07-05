package cache

import (
	"container/list"

	"github.com/ip-api/cache/structs"
	"github.com/ip-api/cache/util"
)

// Based on https://raw.githubusercontent.com/hashicorp/golang-lru/master/simplelru/lru.go
type Cache struct {
	evictList *list.List
	items     map[string]*list.Element
	sizeBytes int
	maxBytes  int
}

// entry is used to hold a value in the evictList
type entry struct {
	key   string
	value *structs.CacheEntry
	size  int
}

func New(maxBytes int) *Cache {
	c := &Cache{
		maxBytes:  maxBytes,
		evictList: list.New(),
		items:     make(map[string]*list.Element),
	}
	return c
}

// Size returns the current size of the cache in bytes.
func (c *Cache) Size() int {
	return c.sizeBytes
}

// addSize adds the specified size to the total cache size and evicts items if it exeeds maxBytes.
func (c *Cache) addSize(size int) {
	c.sizeBytes += size
	for c.sizeBytes > c.maxBytes {
		c.removeOldest()
	}
}

// Add adds a value to the cache.  Returns true if an eviction occurred.
func (c *Cache) Add(key string, value *structs.CacheEntry) {
	size := value.Size()

	// Check for existing item
	if ent, ok := c.items[key]; ok {
		c.evictList.MoveToFront(ent)
		entr := ent.Value.(*entry)
		entr.value = value
		c.sizeBytes -= entr.size
		c.addSize(size)
		return
	}

	// Add new item
	ent := &entry{key, value, size}
	entry := c.evictList.PushFront(ent)
	c.items[key] = entry

	c.addSize(size)
}

// Get looks up a key's value from the cache.
func (c *Cache) Get(key string) *structs.CacheEntry {
	if ent, ok := c.items[key]; ok {
		if ent.Value.(*entry) == nil {
			return nil
		}
		e := ent.Value.(*entry).value
		if e.Expires.Before(util.Now()) {
			return nil
		}
		c.evictList.MoveToFront(ent)
		return e
	}
	return nil
}

// removeOldest removes the oldest item from the cache.
func (c *Cache) removeOldest() {
	ent := c.evictList.Back()
	if ent != nil {
		c.removeElement(ent)
	}
}

// removeElement is used to remove a given list element from the cache
func (c *Cache) removeElement(e *list.Element) {
	c.evictList.Remove(e)
	kv := e.Value.(*entry)
	delete(c.items, kv.key)
	c.sizeBytes -= kv.size
}
