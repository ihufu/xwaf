package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/xwaf/rule_engine/internal/model"
	"github.com/xwaf/rule_engine/internal/repository"
)

const (
	ruleKeyPrefix = "waf:rule:"
	ruleExpire    = 24 * time.Hour
)

// redisCache Redis缓存实现
type redisCache struct {
	client *redis.Client
}

// NewCacheRepository 创建缓存仓储
func NewCacheRepository(client *redis.Client) repository.CacheRepository {
	return &redisCache{
		client: client,
	}
}

// Set 设置缓存
func (c *redisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, data, expiration).Err()
}

// Get 获取缓存
func (c *redisCache) Get(ctx context.Context, key string, value interface{}) error {
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, value)
}

// Delete 删除缓存
func (c *redisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

// Exists 检查缓存是否存在
func (c *redisCache) Exists(ctx context.Context, key string) (bool, error) {
	n, err := c.client.Exists(ctx, key).Result()
	return n > 0, err
}

// GetLock 获取分布式锁
func (c *redisCache) GetLock(key string, expiration time.Duration) repository.Lock {
	return &redisLock{
		client:     c.client,
		key:        "lock:" + key,
		expiration: expiration,
	}
}

// redisLock Redis分布式锁实现
type redisLock struct {
	client     *redis.Client
	key        string
	expiration time.Duration
}

// Lock 获取锁
func (l *redisLock) Lock() bool {
	return l.client.SetNX(context.Background(), l.key, 1, l.expiration).Val()
}

// Unlock 释放锁
func (l *redisLock) Unlock() error {
	return l.client.Del(context.Background(), l.key).Err()
}

// Pipeline 获取管道
func (c *redisCache) Pipeline() repository.Pipeline {
	return newRedisPipeline(c.client.Pipeline())
}

// Publish 发布消息
func (c *redisCache) Publish(ctx context.Context, channel string, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	return c.client.Publish(ctx, channel, data).Err()
}

// SetRule 设置规则缓存
func (c *redisCache) SetRule(ctx context.Context, rule *model.Rule) error {
	key := fmt.Sprintf("%s%d", ruleKeyPrefix, rule.ID)
	return c.Set(ctx, key, rule, ruleExpire)
}

// GetRule 获取规则缓存
func (c *redisCache) GetRule(ctx context.Context, id int64) (*model.Rule, error) {
	key := fmt.Sprintf("%s%d", ruleKeyPrefix, id)
	var rule model.Rule
	err := c.Get(ctx, key, &rule)
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// DeleteRule 删除规则缓存
func (c *redisCache) DeleteRule(ctx context.Context, id int64) error {
	key := fmt.Sprintf("%s%d", ruleKeyPrefix, id)
	return c.Delete(ctx, key)
}

// ClearRules 清空规则缓存
func (c *redisCache) ClearRules(ctx context.Context) error {
	pattern := fmt.Sprintf("%s*", ruleKeyPrefix)
	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := c.Delete(ctx, iter.Val()); err != nil {
			return err
		}
	}
	return iter.Err()
}
