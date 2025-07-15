package redis

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	readType  = "r"
	writeType = "w"
)

// RWClient 读写分离Redis客户端
type RWClient struct {
	master *redis.Client            // 主节点（写操作）
	slaves map[string]*redis.Client // 从节点（读操作），键为节点ID
}

// NewRWClient 创建读写分离Redis客户端
func NewRWClient(redisProps *RedisProperties) (*RWClient, error) {
	var master *redis.Client
	slaves := make(map[string]*redis.Client)

	for _, node := range redisProps.Nodes {
		opts := &redis.Options{
			Addr:         fmt.Sprintf("%s:%d", node.Host, node.Port),
			Password:     node.Password,
			DB:           0, // 默认DB为0，可以根据需要调整
			DialTimeout:  node.Timeout * time.Millisecond,
			MinIdleConns: node.PoolMinIdle,
			PoolSize:     node.PoolMaxActive,
			PoolTimeout:  node.PoolMaxWait * time.Millisecond,
			ReadTimeout:  node.Timeout * time.Millisecond,
			WriteTimeout: node.Timeout * time.Millisecond,
		}

		client := redis.NewClient(opts)
		if err := client.Ping(context.Background()).Err(); err != nil {
			return nil, fmt.Errorf("无法连接到 Redis 节点 %s: %w", node.ID, err)
		}

		switch node.Type {
		case writeType:
			if master != nil {
				return nil, fmt.Errorf("配置了多个写节点，只允许一个")
			}
			master = client
		case readType:
			slaves[node.ID] = client
		default:
			return nil, fmt.Errorf("无效的节点类型: %s", node.Type)
		}
	}

	if master == nil {
		return nil, fmt.Errorf("未配置写节点")
	}

	return &RWClient{
		master: master,
		slaves: slaves,
	}, nil
}

// getRandomSlave 获取一个随机的从节点
func (c *RWClient) getRandomSlave() *redis.Client {
	if len(c.slaves) == 0 {
		return c.master
	}
	var keys []string
	for k := range c.slaves {
		keys = append(keys, k)
	}
	return c.slaves[keys[rand.Intn(len(keys))]]
}

// ====================== 写操作（使用主节点） ======================

func (c *RWClient) Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd {
	return c.master.Set(ctx, key, value, expiration)
}

func (c *RWClient) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	return c.master.Del(ctx, keys...)
}

func (c *RWClient) HSet(ctx context.Context, key string, values ...any) *redis.IntCmd {
	return c.master.HSet(ctx, key, values...)
}

func (c *RWClient) Incr(ctx context.Context, key string) *redis.IntCmd {
	return c.master.Incr(ctx, key)
}

// ====================== 读操作（使用从节点） ======================

func (c *RWClient) Get(ctx context.Context, key string) *redis.StringCmd {
	slave := c.getRandomSlave()
	return slave.Get(ctx, key)
}

func (c *RWClient) HGet(ctx context.Context, key, field string) *redis.StringCmd {
	slave := c.getRandomSlave()
	return slave.HGet(ctx, key, field)
}

func (c *RWClient) Exists(ctx context.Context, keys ...string) *redis.IntCmd {
	slave := c.getRandomSlave()
	return slave.Exists(ctx, keys...)
}

func (c *RWClient) Keys(ctx context.Context, pattern string) *redis.StringSliceCmd {
	slave := c.getRandomSlave()
	return slave.Keys(ctx, pattern)
}

// ====================== 管理方法 ======================

// Ping 检查所有节点连接
func (c *RWClient) Ping(ctx context.Context) error {
	// 检查主节点
	if err := c.master.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("主节点 ping 失败: %w", err)
	}

	// 检查从节点
	for id, slave := range c.slaves {
		if err := slave.Ping(ctx).Err(); err != nil {
			return fmt.Errorf("从节点 %s ping 失败: %w", id, err)
		}
	}

	return nil
}

// Close 关闭所有连接
func (c *RWClient) Close() error {
	masterErr := c.master.Close()

	var slaveErrs []error
	for id, slave := range c.slaves {
		if err := slave.Close(); err != nil {
			slaveErrs = append(slaveErrs, fmt.Errorf("从节点 %s 关闭失败: %w", id, err))
		}
	}

	if masterErr != nil {
		slaveErrs = append(slaveErrs, masterErr)
	}

	if len(slaveErrs) > 0 {
		errMsg := "关闭客户端时出错:"
		for _, err := range slaveErrs {
			errMsg += fmt.Sprintf("\n - %v", err)
		}
		return errors.New(errMsg)
	}

	return nil
}

// Master 获取主节点客户端（用于直接访问）
func (c *RWClient) Master() *redis.Client {
	return c.master
}

// Slaves 获取从节点客户端列表（用于直接访问）
func (c *RWClient) Slaves() map[string]*redis.Client {
	return c.slaves
}
