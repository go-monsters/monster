package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-monsters/monster/internals/logs/merror"
	"github.com/go-monsters/monster/pkg/cache"
)

var DefaultEvery = 60 // 1 minute

type Cache struct {
	sync.RWMutex
	dur   time.Duration
	items map[string]*Item
	Every int // run an expiration check Every clock time
}

func NewMemoryCache() cache.Cache {
	return &Cache{items: make(map[string]*Item)}
}

func (c *Cache) GetClient() interface{} {
	return c
}

func (c *Cache) Get(ctx context.Context, key string) (interface{}, error) {
	c.RLock()
	defer c.RUnlock()
	if itm, ok := c.items[key]; ok {
		if itm.isExpire() {
			return nil, cache.ErrKeyExpired
		}
		return itm.val, nil
	}
	return nil, cache.ErrKeyNotExist
}

func (c *Cache) GetMulti(ctx context.Context, keys []string) ([]interface{}, error) {
	rc := make([]interface{}, len(keys))
	keysErr := make([]string, 0)

	for i, ki := range keys {
		val, err := c.Get(ctx, ki)
		if err != nil {
			keysErr = append(keysErr, fmt.Sprintf("key [%s] error: %s", ki, err.Error()))
			continue
		}
		rc[i] = val
	}

	if len(keysErr) == 0 {
		return rc, nil
	}
	return rc, merror.Error(strings.Join(keysErr, "; "))
}

func (c *Cache) Put(ctx context.Context, key string, val interface{}, timeout time.Duration) error {
	c.Lock()
	defer c.Unlock()
	c.items[key] = &Item{
		val:         val,
		createdTime: time.Now(),
		lifespan:    timeout,
	}
	return nil
}

func (c *Cache) Delete(ctx context.Context, key string) error {
	c.Lock()
	defer c.Unlock()
	delete(c.items, key)
	return nil
}

func (c *Cache) Start(config string) error {
	var cf map[string]int
	if err := json.Unmarshal([]byte(config), &cf); err != nil {
		return merror.Wrapf(err, "invalid config, please check your input: %s", config)
	}
	if _, ok := cf["interval"]; !ok {
		cf = make(map[string]int)
		cf["interval"] = DefaultEvery
	}
	dur := time.Duration(cf["interval"]) * time.Second
	c.Every = cf["interval"]
	c.dur = dur
	go c.vacuum()
	return nil
}

func (c *Cache) vacuum() {
	c.RLock()
	every := c.Every
	c.RUnlock()

	if every < 1 {
		return
	}
	for {
		<-time.After(c.dur)
		c.RLock()
		if c.items == nil {
			c.RUnlock()
			return
		}
		c.RUnlock()
		if keys := c.expiredKeys(); len(keys) != 0 {
			c.clearItems(keys)
		}
	}
}

func (c *Cache) expiredKeys() (keys []string) {
	c.RLock()
	defer c.RUnlock()
	for key, itm := range c.items {
		if itm.isExpire() {
			keys = append(keys, key)
		}
	}
	return
}

func (c *Cache) clearItems(keys []string) {
	c.Lock()
	defer c.Unlock()
	for _, key := range keys {
		delete(c.items, key)
	}
}

func init() {
	cache.RegisterNewCacheImpl("memory", NewMemoryCache)
}
