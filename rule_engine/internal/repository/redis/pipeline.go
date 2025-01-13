package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/xwaf/rule_engine/internal/repository"
)

// redisPipeline Redis管道实现
type redisPipeline struct {
	pipe redis.Pipeliner
}

// newRedisPipeline 创建Redis管道
func newRedisPipeline(pipe redis.Pipeliner) repository.Pipeline {
	return &redisPipeline{
		pipe: pipe,
	}
}

// Set 设置缓存
func (p *redisPipeline) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) {
	data, _ := json.Marshal(value)
	p.pipe.Set(ctx, key, data, expiration)
}

// Exec 执行管道命令
func (p *redisPipeline) Exec(ctx context.Context) ([]interface{}, error) {
	cmds, err := p.pipe.Exec(ctx)
	if err != nil {
		return nil, err
	}

	results := make([]interface{}, len(cmds))
	for i, cmd := range cmds {
		results[i] = cmd
	}
	return results, nil
}
