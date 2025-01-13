package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/xwaf/rule_engine/internal/model"
	"github.com/xwaf/rule_engine/internal/repository"
)

// redisRuleCache Redis规则缓存实现
type redisRuleCache struct {
	client *redis.Client
}

// NewRedisRuleCache 创建Redis规则缓存
func NewRedisRuleCache(client *redis.Client) repository.RuleCache {
	return &redisRuleCache{
		client: client,
	}
}

// SetRule 设置规则缓存
func (c *redisRuleCache) SetRule(ctx context.Context, rule *model.Rule) error {
	data, err := json.Marshal(rule)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("%s%d", ruleKeyPrefix, rule.ID)
	return c.client.Set(ctx, key, data, ruleExpire).Err()
}

// GetRule 获取规则缓存
func (c *redisRuleCache) GetRule(ctx context.Context, id int64) (*model.Rule, error) {
	key := fmt.Sprintf("%s%d", ruleKeyPrefix, id)
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var rule model.Rule
	if err := json.Unmarshal(data, &rule); err != nil {
		return nil, err
	}

	return &rule, nil
}

// DeleteRule 删除规则缓存
func (c *redisRuleCache) DeleteRule(ctx context.Context, id int64) error {
	key := fmt.Sprintf("%s%d", ruleKeyPrefix, id)
	return c.client.Del(ctx, key).Err()
}

// ClearRules 清空规则缓存
func (c *redisRuleCache) ClearRules(ctx context.Context) error {
	pattern := fmt.Sprintf("%s*", ruleKeyPrefix)
	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()

	for iter.Next(ctx) {
		if err := c.client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}

	return iter.Err()
}
