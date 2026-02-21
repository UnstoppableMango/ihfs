package tarfs

import (
	"strings"
	"sync"
)

type cache struct {
	mux  sync.RWMutex
	data map[string]*fileData
}

func (c *cache) get(name string) *fileData {
	c.mux.RLock()
	defer c.mux.RUnlock()
	return c.data[name]
}

func (c *cache) set(name string, fd *fileData) {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.data[name] = fd
}

func (c *cache) hasPrefix(prefix string) bool {
	c.mux.RLock()
	defer c.mux.RUnlock()

	for name := range c.data {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}

func (c *cache) listWithPrefix(prefix string) []*fileData {
	c.mux.RLock()
	defer c.mux.RUnlock()

	var result []*fileData
	for name, fd := range c.data {
		// Empty prefix matches all (for root directory)
		if prefix == "" || strings.HasPrefix(name, prefix) {
			result = append(result, fd)
		}
	}
	return result
}

func newCache() *cache {
	return &cache{
		mux:  sync.RWMutex{},
		data: make(map[string]*fileData),
	}
}
