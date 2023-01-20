package file

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/go-monsters/monster/internals/logs/merror"
	"github.com/go-monsters/monster/pkg/cache"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var (
	FileCachePath           = "cache"     // cache directory
	FileCacheFileSuffix     = ".bin"      // cache file suffix
	FileCacheDirectoryLevel = 2           // cache file deep level if auto generated cache files.
	FileCacheEmbedExpiry    time.Duration // cache expire time, default is no expire forever.
)

type Cache struct {
	CachePath      string
	FileSuffix     string
	DirectoryLevel int
	EmbedExpiry    int
}

func NewFileCache() cache.Cache {
	return &Cache{}
}

func (c *Cache) Get(ctx context.Context, key string) (interface{}, error) {
	fn, err := c.getCacheFileName(key)
	if err != nil {
		return nil, err
	}
	fileData, err := fileGetContents(fn)
	if err != nil {
		return nil, err
	}

	var to Item
	err = GobDecode(fileData, &to)
	if err != nil {
		return nil, err
	}

	if to.Expired.Before(time.Now()) {
		return nil, cache.ErrKeyExpired
	}
	return to.Data, nil
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
	gob.Register(val)

	item := Item{Data: val}
	if timeout == time.Duration(c.EmbedExpiry) {
		item.Expired = time.Now().Add((86400 * 365 * 10) * time.Second) // ten years
	} else {
		item.Expired = time.Now().Add(timeout)
	}
	item.LastAccess = time.Now()
	data, err := GobEncode(item)
	if err != nil {
		return err
	}

	fn, err := c.getCacheFileName(key)
	if err != nil {
		return err
	}
	return PutContents(fn, data)
}

func (c *Cache) Delete(ctx context.Context, key string) error {
	filename, err := c.getCacheFileName(key)
	if err != nil {
		return err
	}
	if ok, _ := exists(filename); ok {
		err = os.Remove(filename)
		if err != nil {
			return merror.Wrapf(err,
				"can not delete this file cache key-value, key is %s and file name is %s", key, filename)
		}
	}
	return nil
}

func (c *Cache) Start(config string) error {
	cfg := make(map[string]string)
	err := json.Unmarshal([]byte(config), &cfg)
	if err != nil {
		return err
	}

	const cpKey = "CachePath"
	const fsKey = "FileSuffix"
	const dlKey = "DirectoryLevel"
	const eeKey = "EmbedExpiry"

	if _, ok := cfg[cpKey]; !ok {
		cfg[cpKey] = FileCachePath
	}

	if _, ok := cfg[fsKey]; !ok {
		cfg[fsKey] = FileCacheFileSuffix
	}

	if _, ok := cfg[dlKey]; !ok {
		cfg[dlKey] = strconv.Itoa(FileCacheDirectoryLevel)
	}

	if _, ok := cfg[eeKey]; !ok {
		cfg[eeKey] = strconv.FormatInt(int64(FileCacheEmbedExpiry.Seconds()), 10)
	}
	c.CachePath = cfg[cpKey]
	c.FileSuffix = cfg[fsKey]
	c.DirectoryLevel, err = strconv.Atoi(cfg[dlKey])
	if err != nil {
		return merror.Wrapf(err,
			"invalid directory level config, please check your input, it must be integer: %s", cfg[dlKey])
	}
	c.EmbedExpiry, err = strconv.Atoi(cfg[eeKey])
	if err != nil {
		return merror.Wrapf(err,
			"invalid embed expiry config, please check your input, it must be integer: %s", cfg[eeKey])
	}
	return c.Create()
}

func (c *Cache) Create() error {
	ok, err := exists(c.CachePath)
	if err != nil || ok {
		return err
	}
	err = os.MkdirAll(c.CachePath, os.ModePerm)
	if err != nil {
		return merror.Wrapf(err,
			"could not create directory, please check the config [%s] and file mode.", c.CachePath)
	}
	return nil
}

func (c *Cache) getCacheFileName(key string) (string, error) {
	m := md5.New()
	_, _ = io.WriteString(m, key)
	keyMd5 := hex.EncodeToString(m.Sum(nil))
	cachePath := c.CachePath
	switch c.DirectoryLevel {
	case 2:
		cachePath = filepath.Join(cachePath, keyMd5[0:2], keyMd5[2:4])
	case 1:
		cachePath = filepath.Join(cachePath, keyMd5[0:2])
	}
	ok, err := exists(cachePath)
	if err != nil {
		return "", err
	}
	if !ok {
		err = os.MkdirAll(cachePath, os.ModePerm)
		if err != nil {
			return "", merror.Wrapf(err,
				"could not create the directory: %s", cachePath)
		}
	}

	return filepath.Join(cachePath, fmt.Sprintf("%s%s", keyMd5, c.FileSuffix)), nil
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, merror.Wrapf(err, "file cache path is invalid: %s", path)
}

func fileGetContents(filename string) ([]byte, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, merror.Wrapf(err,
			"could not read the data from the file: %s, "+
				"please confirm that file exist and monster has the permission to read the content.", filename)
	}
	return data, nil
}

func GobEncode(data interface{}) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(data)
	if err != nil {
		return nil, merror.Wrap(err, "could not encode this data")
	}
	return buf.Bytes(), nil
}

// GobDecode Gob decodes a file cache item.
func GobDecode(data []byte, to *Item) error {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&to)
	if err != nil {
		return merror.Wrap(err,
			"could not decode this data to FileCacheItem. Make sure that the data is encoded by GOB.")
	}
	return nil
}

func PutContents(filename string, content []byte) error {
	return os.WriteFile(filename, content, os.ModePerm)
}

func init() {
	cache.RegisterNewCacheImpl("file", NewFileCache)
}
