package cache

import (
	"context"
	"time"

	"github.com/go-monsters/monster/internals/logs/merror"
)

var (
	ErrKeyExpired  = merror.Error("the key is expired")
	ErrKeyNotExist = merror.Error("the key isn't exist")
)

type Cache interface {
	Get(ctx context.Context, key string) (interface{}, error)

	GetMulti(ctx context.Context, keys []string) ([]interface{}, error)

	Put(ctx context.Context, key string, val interface{}, timeout time.Duration) error

	Delete(ctx context.Context, key string) error

	Start(config string) error
}

type Instance func() Cache

var impls = make(map[string]Instance)

func RegisterNewCacheImpl(name string, impl Instance) {
	if impl == nil {
		panic(merror.Error("cache: Register adapter is nil").Error())
	}
	if _, ok := impls[name]; ok {
		panic("cache: Register called twice for adapter " + name)
	}
	impls[name] = impl
}

func NewCache(implName, config string) (adapter Cache, err error) {
	instanceFunc, ok := impls[implName]
	if !ok {
		err = merror.Errorf("cache: unknown impl name %s (forgot to import?)", implName)
		return
	}
	adapter = instanceFunc()
	err = adapter.Start(config)
	if err != nil {
		adapter = nil
	}
	return
}
