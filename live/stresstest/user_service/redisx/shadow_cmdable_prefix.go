package redisx

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v9"
	"time"
)

type ShadowCmdablePrefix struct {
	redis.Cmdable
	prefix string
}

func NewShadowCmdablePrefix(cmd redis.Cmdable) *ShadowCmdablePrefix {
	return &ShadowCmdablePrefix{
		prefix:  "shadow",
		Cmdable: cmd,
	}
}

func (s *ShadowCmdablePrefix) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	for idx, k := range keys {
		keys[idx] = fmt.Sprintf("%s:%s", s.prefix, k)
	}
	return s.Cmdable.Del(ctx, keys...)
}

func (s *ShadowCmdablePrefix) Set(ctx context.Context, key string,
	value interface{}, expiration time.Duration) *redis.StatusCmd {
	key = fmt.Sprintf("%s:%s", s.prefix, key)
	return s.Cmdable.Set(ctx, key, value, expiration)
}

func (s *ShadowCmdablePrefix) Get(ctx context.Context, key string) *redis.StringCmd {
	key = fmt.Sprintf("%s:%s", s.prefix, key)
	return s.Cmdable.Get(ctx, key)
}
