package redis

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/frochyzhang/ag-core/ag/ag_cache"
)

// RedisCache 使用 Redis 实现 ICache 接口
type RedisCache[T any] struct {
	Client *RWClient
}

// Get 获取 key 对应的值
func (c *RedisCache[T]) Get(key string) (T, bool) {
	cmd := c.Client.Get(context.Background(), key)
	if cmd.Err() != nil {
		var zero T
		return zero, false
	}
	strValue, _ := cmd.Result()
	// 使用 json.Unmarshal 将字符串转换为 T 类型
	var value T
	if err := json.Unmarshal([]byte(strValue), &value); err != nil {
		var zero T
		return zero, false
	}
	return value, true
}

// GetWithCustLoader 如果 key 不存在则调用 loader 加载
func (c *RedisCache[T]) GetWithCustLoader(ctx context.Context, key string, loader func(context.Context, string) (T, error)) (T, error) {
	v, ok := c.Get(key)
	if !ok {
		newValue, err := loader(ctx, key)
		if err != nil {
			return newValue, err
		}
		c.Set(key, newValue) // 默认不过期
		return newValue, nil
	}
	return v, nil
}

// GetWithLoader 如果 key 不存在则调用默认加载器加载
func (c *RedisCache[T]) GetWithLoader(ctx context.Context, key string) (T, error) {
	// 这里应该使用 Config.DefaultLoader 来代替硬编码的 loader 函数
	panic("GetWithLoader 方法未实现，请根据实际情况填充")
}

// Set 设置 key-value 对
func (c *RedisCache[T]) Set(key string, value T) error {
	cmd := c.Client.Set(context.Background(), key, value, 0) // 默认不过期
	return cmd.Err()
}

// SetWithExpire 设置带过期时间的 key-value 对
func (c *RedisCache[T]) SetWithExpire(key string, value T, ttl int64) error {
	return c.SetWithExpireInternal(key, value, time.Duration(ttl)*time.Second)
}

// SetWithExpireInternal 内部方法，用于设置带过期时间的 key-value 对
func (c *RedisCache[T]) SetWithExpireInternal(key string, value T, expiration time.Duration) error {
	cmd := c.Client.Set(context.Background(), key, value, expiration)
	return cmd.Err()
}

// Del 删除指定 key
func (c *RedisCache[T]) Del(key string) error {
	cmd := c.Client.Del(context.Background(), key)
	return cmd.Err()
}

// Clear 清空所有缓存 - 注意：这将在整个数据库中执行，非常危险！
func (c *RedisCache[T]) Clear() error {
	return errors.New("清空操作不安全且不可逆，已禁用")
}

// Stats 获取缓存统计信息 - Redis 不直接支持此功能，因此返回一个默认的结构体
func (c *RedisCache[T]) Stats() cache.Stats {
	return cache.Stats{}
}
