package redisx

import (
	"context"
	"github.com/go-redis/redis/v9"
	"sync"
)

type ShadowCmdable struct {
	live   redis.Cmdable
	shadow redis.Cmdable
}

func (s *ShadowCmdable) Pipeline() redis.Pipeliner {
	//return &shadowPipeline{
	//	cmd: s,
	//}
	return &doubleShadowPipeline{
		Pipeliner:  s.live.Pipeline(),
		shadowPipe: s.shadow.Pipeline(),
	}
}

// 允许命令一会在 live 上，一会在 shadow 上
type doubleShadowPipeline struct {
	redis.Pipeliner
	shadowPipe redis.Pipeliner
}

func (s *doubleShadowPipeline) Ping(ctx context.Context) *redis.StatusCmd {
	if isShadow(ctx) {
		return s.shadowPipe.Ping(ctx)
	}
	return s.Pipeliner.Ping(ctx)
}

// 这个要求后续所有的命令，要么都在 live 上，要么都在 shadow 上。
type shadowPipeline struct {
	redis.Pipeliner
	cmd  *ShadowCmdable
	once sync.Once
}

func (s *shadowPipeline) Ping(ctx context.Context) *redis.StatusCmd {
	s.once.Do(func() {
		if isShadow(ctx) {
			s.Pipeliner = s.cmd.shadow.Pipeline()
			return
		}
		s.Pipeliner = s.cmd.live.Pipeline()
	})
	return s.Pipeliner.Ping(ctx)
}

func isShadow(ctx context.Context) bool {
	return ctx.Value("stress-test") == "true"
}
