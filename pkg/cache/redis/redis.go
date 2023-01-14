package redis

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/go-monsters/monster/internals/logs/merror"
	"github.com/go-monsters/monster/pkg/cache"
	"github.com/go-redis/redis"
	"go.elastic.co/apm/module/apmgoredis/v2"
)

var DefaultKey = "monsterCacheRedis"

type Cache struct {
	conn     apmgoredis.Client
	connInfo string
	dbNum    int
	key      string
	password string
	minIdle  int
}

func NewRedisCache() cache.Cache {
	return &Cache{key: DefaultKey}
}

func (c *Cache) Get(ctx context.Context, key string) (interface{}, error) {
	client := c.conn.WithContext(ctx)
	res := client.Get(key)
	if res.Err() != nil {
		return nil, merror.Wrapf(res.Err(), "error with get")
	}
	return res.Val(), nil
}

func (c *Cache) GetMulti(ctx context.Context, keys []string) ([]interface{}, error) {
	client := c.conn.WithContext(ctx)
	res := client.MGet(keys...)
	if res.Err() != nil {
		return nil, res.Err()
	}
	values := make([]interface{}, 0)
	j, err := json.Marshal(res.Val())
	if err != nil {
		return nil, res.Err()
	}
	err = json.Unmarshal(j, &values)
	if err != nil {
		return nil, res.Err()
	}
	return values, nil
}

func (c *Cache) Put(ctx context.Context, key string, val interface{}, timeout time.Duration) error {
	conn := c.conn.WithContext(ctx)
	res := conn.Set(key, val, timeout)
	if res.Err() != nil {
		return res.Err()
	}
	return nil
}

func (c *Cache) Delete(ctx context.Context, key string) error {
	conn := c.conn.WithContext(ctx)
	res := conn.Del(key)
	if res.Err() != nil {
		return res.Err()
	}
	return nil
}

func (c *Cache) Start(config string) error {
	var cf map[string]string
	err := json.Unmarshal([]byte(config), &cf)
	if err != nil {
		return merror.Wrapf(err, "could not unmarshal the config: %s", config)
	}
	if _, ok := cf["key"]; !ok {
		cf["key"] = DefaultKey
	}
	if _, ok := cf["conn"]; !ok {
		return merror.Wrapf(err, "config missing conn field: %s", config)
	}

	// Format redis://<password>@<host>:<port>
	cf["conn"] = strings.Replace(cf["conn"], "redis://", "", 1)
	if i := strings.Index(cf["conn"], "@"); i > -1 {
		cf["password"] = cf["conn"][0:i]
		cf["conn"] = cf["conn"][i+1:]
	}

	if _, ok := cf["dbNum"]; !ok {
		cf["dbNum"] = "0"
	}
	if _, ok := cf["password"]; !ok {
		cf["password"] = ""
	}
	if _, ok := cf["minIdle"]; !ok {
		cf["minIdle"] = "3"
	}

	c.key = cf["key"]
	c.connInfo = cf["conn"]
	c.dbNum, _ = strconv.Atoi(cf["dbNum"])
	c.password = cf["password"]
	c.minIdle, _ = strconv.Atoi(cf["minIdle"])

	c.connectInit()

	conn := c.conn.RedisClient()
	defer func() {
		_ = conn.Close()
	}()

	return nil
}

func (c *Cache) connectInit() {
	client := redis.NewClient(&redis.Options{
		Addr:         c.connInfo,
		Password:     c.password,
		DB:           c.dbNum,
		MinIdleConns: c.minIdle,
	})
	c.conn = apmgoredis.Wrap(client)
}

func init() {
	cache.RegisterNewCacheImpl("redis", NewRedisCache)
}
