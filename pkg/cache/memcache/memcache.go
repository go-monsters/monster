package memcache

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/bradfitz/gomemcache/memcache"

	"github.com/go-monsters/monster/internals/logs/merror"
	"github.com/go-monsters/monster/pkg/cache"
)

type Cache struct {
	conn     *memcache.Client
	connInfo []string
}

func NewMemCache() cache.Cache {
	return &Cache{}
}

func (c Cache) Get(ctx context.Context, key string) (interface{}, error) {
	if item, err := c.conn.Get(key); err == nil {
		return item.Value, nil
	} else {
		return nil, merror.Wrapf(err,
			"could not read data from memcache, please check your key, network and connection. Root cause: %s",
			err.Error())
	}
}

func (c Cache) GetMulti(ctx context.Context, keys []string) ([]interface{}, error) {
	rv := make([]interface{}, len(keys))

	mv, err := c.conn.GetMulti(keys)
	if err != nil {
		return rv, merror.Wrapf(err,
			"could not read multiple key-values from memcache, "+
				"please check your keys, network and connection. Root cause: %s",
			err.Error())
	}

	keysErr := make([]string, 0)
	for i, ki := range keys {
		if _, ok := mv[ki]; !ok {
			keysErr = append(keysErr, fmt.Sprintf("key [%s] error: %s", ki, "key not exist"))
			continue
		}
		rv[i] = mv[ki].Value
	}

	if len(keysErr) == 0 {
		return rv, nil
	}
	return rv, merror.Error(strings.Join(keysErr, "; "))
}

func (c Cache) Put(ctx context.Context, key string, val interface{}, timeout time.Duration) error {
	item := memcache.Item{Key: key, Expiration: int32(timeout / time.Second)}
	if v, ok := val.([]byte); ok {
		item.Value = v
	} else if str, ok := val.(string); ok {
		item.Value = []byte(str)
	} else {
		return merror.Errorf("the value must be string or byte[]. key: %s, value:%v", key, val)
	}
	return merror.Wrapf(c.conn.Set(&item),
		"could not put key-value to memcache, key: %s", key)
}

func (c Cache) Delete(ctx context.Context, key string) error {
	return merror.Wrapf(c.conn.Delete(key),
		"could not delete key-value from memcache, key: %s", key)
}

func (c Cache) Start(config string) error {
	var cf map[string]string
	if err := json.Unmarshal([]byte(config), &cf); err != nil {
		return merror.Wrapf(err,
			"could not unmarshal this config, it must be valid json stringP: %s", config)
	}

	if _, ok := cf["conn"]; !ok {
		return merror.Errorf(`config must contains "conn" field: %s`, config)
	}
	c.connInfo = strings.Split(cf["conn"], ";")
	c.conn = memcache.New(c.connInfo...)
	return nil
}

func init() {
	cache.RegisterNewCacheImpl("memcache", NewMemCache)
}
