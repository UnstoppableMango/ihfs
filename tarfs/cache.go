package tarfs

import "sync"

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

func newCache() *cache {
	return &cache{
		mux:  sync.RWMutex{},
		data: make(map[string]*fileData),
	}
}
