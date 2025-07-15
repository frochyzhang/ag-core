package fxs

import (
	"context"

	ag_redis "github.com/frochyzhang/ag-core/ag/ag_cache/redis"
	"github.com/frochyzhang/ag-core/ag/ag_conf"
	"go.uber.org/fx"
)

// FxRedisServerMode 是一个 Fx 模块，用于初始化 Redis 客户端并注册生命周期钩子
var FxRedisServerMode = fx.Module("redis",
	fx.Provide(
		FxBuildRedisProperties,
		ag_redis.NewRWClient, // 使用读写分离客户端
		NewRedisCache,        // 提供 RedisCache 实例
	),
	fx.Invoke(registerHooks),
)

type FxRedisInParam struct {
	fx.In

	Env    ag_conf.IConfigurableEnvironment
	Binder ag_conf.IBinder
}

func FxBuildRedisProperties(params FxRedisInParam) (*ag_redis.RedisProperties, error) {
	builder := ag_redis.NewRedisConfigBuilder(params.Binder)
	return builder.BuildConfig()
}

// NewRedisCache 提供 RedisCache 实例
func NewRedisCache(client *ag_redis.RWClient) *ag_redis.RedisCache[any] {
	return &ag_redis.RedisCache[any]{Client: client}
}

// registerHooks 注册生命周期钩子
func registerHooks(lc fx.Lifecycle, client *ag_redis.RWClient) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// 启动时健康检查
			return client.Ping(ctx)
		},
		OnStop: func(ctx context.Context) error {
			// 关闭连接
			return client.Close()
		},
	})
}
