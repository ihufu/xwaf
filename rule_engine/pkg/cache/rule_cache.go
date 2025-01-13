package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/patrickmn/go-cache"
	"github.com/xwaf/rule_engine/internal/errors"
	"github.com/xwaf/rule_engine/pkg/metrics"
)

// RuleCache 规则缓存接口
type RuleCache interface {
	// Get 获取规则
	Get(ctx context.Context, key string) (interface{}, error)
	// Set 设置规则
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	// Delete 删除规则
	Delete(ctx context.Context, key string) error
	// Clear 清空缓存
	Clear(ctx context.Context) error
}

// LocalCache 本地内存缓存
type LocalCache struct {
	cache *cache.Cache
}

// NewLocalCache 创建本地缓存
func NewLocalCache(defaultExpiration, cleanupInterval time.Duration) *LocalCache {
	return &LocalCache{
		cache: cache.New(defaultExpiration, cleanupInterval),
	}
}

func (c *LocalCache) Get(_ context.Context, key string) (interface{}, error) {
	start := time.Now()
	defer func() {
		metrics.CacheLatency.WithLabelValues("get", "local").Observe(time.Since(start).Seconds())
	}()

	if value, found := c.cache.Get(key); found {
		metrics.CacheHits.WithLabelValues("local").Inc()
		return value, nil
	}
	metrics.CacheMisses.WithLabelValues("local").Inc()
	return nil, errors.NewError(errors.ErrCacheMiss, fmt.Sprintf("key not found: %s", key))
}

func (c *LocalCache) Set(_ context.Context, key string, value interface{}, expiration time.Duration) error {
	start := time.Now()
	defer func() {
		metrics.CacheLatency.WithLabelValues("set", "local").Observe(time.Since(start).Seconds())
	}()

	c.cache.Set(key, value, expiration)
	return nil
}

func (c *LocalCache) Delete(_ context.Context, key string) error {
	start := time.Now()
	defer func() {
		metrics.CacheLatency.WithLabelValues("delete", "local").Observe(time.Since(start).Seconds())
	}()

	c.cache.Delete(key)
	return nil
}

func (c *LocalCache) Clear(_ context.Context) error {
	start := time.Now()
	defer func() {
		metrics.CacheLatency.WithLabelValues("clear", "local").Observe(time.Since(start).Seconds())
	}()

	c.cache.Flush()
	return nil
}

// RedisCache Redis缓存
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache 创建Redis缓存
func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{
		client: client,
	}
}

func (c *RedisCache) Get(ctx context.Context, key string) (interface{}, error) {
	start := time.Now()
	defer func() {
		metrics.CacheLatency.WithLabelValues("get", "redis").Observe(time.Since(start).Seconds())
	}()

	value, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			metrics.CacheMisses.WithLabelValues("redis").Inc()
			return nil, errors.NewError(errors.ErrCacheMiss, fmt.Sprintf("key not found: %s", key))
		}
		return nil, errors.NewError(errors.ErrCache, err)
	}

	metrics.CacheHits.WithLabelValues("redis").Inc()
	var result interface{}
	if err := json.Unmarshal([]byte(value), &result); err != nil {
		return nil, errors.NewError(errors.ErrCacheInvalid, err)
	}
	return result, nil
}

func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	start := time.Now()
	defer func() {
		metrics.CacheLatency.WithLabelValues("set", "redis").Observe(time.Since(start).Seconds())
	}()

	data, err := json.Marshal(value)
	if err != nil {
		return errors.NewError(errors.ErrCacheInvalid, err)
	}
	if err := c.client.Set(ctx, key, data, expiration).Err(); err != nil {
		return errors.NewError(errors.ErrCache, err)
	}
	return nil
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
	start := time.Now()
	defer func() {
		metrics.CacheLatency.WithLabelValues("delete", "redis").Observe(time.Since(start).Seconds())
	}()

	if err := c.client.Del(ctx, key).Err(); err != nil {
		return errors.NewError(errors.ErrCache, err)
	}
	return nil
}

func (c *RedisCache) Clear(ctx context.Context) error {
	start := time.Now()
	defer func() {
		metrics.CacheLatency.WithLabelValues("clear", "redis").Observe(time.Since(start).Seconds())
	}()

	if err := c.client.FlushDB(ctx).Err(); err != nil {
		return errors.NewError(errors.ErrCache, err)
	}
	return nil
}

// TwoLevelCache 两级缓存
type TwoLevelCache struct {
	local  RuleCache
	redis  RuleCache
	mutex  sync.RWMutex
	prefix string
}

// NewTwoLevelCache 创建两级缓存
func NewTwoLevelCache(local RuleCache, redis RuleCache, prefix string) *TwoLevelCache {
	return &TwoLevelCache{
		local:  local,
		redis:  redis,
		prefix: prefix,
	}
}

func (c *TwoLevelCache) getKey(key string) string {
	return fmt.Sprintf("%s:%s", c.prefix, key)
}

func (c *TwoLevelCache) Get(ctx context.Context, key string) (interface{}, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	start := time.Now()
	defer func() {
		metrics.CacheLatency.WithLabelValues("get", "two_level").Observe(time.Since(start).Seconds())
	}()

	// 先查本地缓存
	if value, err := c.local.Get(ctx, key); err == nil {
		metrics.CacheHits.WithLabelValues("two_level").Inc()
		return value, nil
	}

	// 本地缓存未命中，查Redis
	value, err := c.redis.Get(ctx, c.getKey(key))
	if err != nil {
		metrics.CacheMisses.WithLabelValues("two_level").Inc()
		return nil, errors.NewError(errors.ErrCache, err)
	}

	// 将Redis中的数据写入本地缓存
	if err := c.local.Set(ctx, key, value, time.Minute*5); err != nil {
		return nil, errors.NewError(errors.ErrCache, err)
	}

	metrics.CacheHits.WithLabelValues("two_level").Inc()
	return value, nil
}

func (c *TwoLevelCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	start := time.Now()
	defer func() {
		metrics.CacheLatency.WithLabelValues("set", "two_level").Observe(time.Since(start).Seconds())
	}()

	// 同时写入本地缓存和Redis
	if err := c.local.Set(ctx, key, value, expiration); err != nil {
		return errors.NewError(errors.ErrCache, err)
	}

	if err := c.redis.Set(ctx, c.getKey(key), value, expiration); err != nil {
		return errors.NewError(errors.ErrCache, err)
	}
	return nil
}

func (c *TwoLevelCache) Delete(ctx context.Context, key string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	start := time.Now()
	defer func() {
		metrics.CacheLatency.WithLabelValues("delete", "two_level").Observe(time.Since(start).Seconds())
	}()

	// 同时删除本地缓存和Redis中的数据
	if err := c.local.Delete(ctx, key); err != nil {
		return errors.NewError(errors.ErrCache, err)
	}

	if err := c.redis.Delete(ctx, c.getKey(key)); err != nil {
		return errors.NewError(errors.ErrCache, err)
	}
	return nil
}

func (c *TwoLevelCache) Clear(ctx context.Context) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	start := time.Now()
	defer func() {
		metrics.CacheLatency.WithLabelValues("clear", "two_level").Observe(time.Since(start).Seconds())
	}()

	// 同时清空本地缓存和Redis
	if err := c.local.Clear(ctx); err != nil {
		return errors.NewError(errors.ErrCache, err)
	}

	if err := c.redis.Clear(ctx); err != nil {
		return errors.NewError(errors.ErrCache, err)
	}
	return nil
}
