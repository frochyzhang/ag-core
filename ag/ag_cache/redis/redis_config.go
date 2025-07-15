package redis

import (
	"fmt"
	"log"
	"net"

	"github.com/frochyzhang/ag-core/ag/ag_conf"
)

const (
	redisPrefix = "redis"
)

// RedisProperties Redis 配置（支持读写分离）
type RedisProperties struct {
	Nodes []RedisNodeConfig `value:"${nodes:[]}"` // Redis 节点列表
}

// RedisConfigBuilder 构建 Redis 配置
type RedisConfigBuilder struct {
	binder ag_conf.IBinder
}

// NewRedisConfigBuilder 创建一个新的 RedisConfigBuilder 实例
func NewRedisConfigBuilder(binder ag_conf.IBinder) *RedisConfigBuilder {
	return &RedisConfigBuilder{
		binder: binder,
	}
}

// BuildConfig 构建 Redis 配置
func (builder *RedisConfigBuilder) BuildConfig() (*RedisProperties, error) {
	var redisProps RedisProperties
	err := builder.binder.Bind(&redisProps, redisPrefix)
	if err != nil {
		log.Printf("加载 Redis 配置失败: %v", err)
		return nil, fmt.Errorf("加载 Redis 配置失败: %w", err)
	}

	for _, node := range redisProps.Nodes {
		if err := validateRedisNode(&node); err != nil {
			return nil, fmt.Errorf("无效的 Redis 配置：%w", err)
		}
	}

	return &redisProps, nil
}

// 验证单个节点
func validateRedisNode(node *RedisNodeConfig) error {
	if node == nil {
		return fmt.Errorf("redis 节点配置为空")
	}

	// 验证地址格式
	addr := fmt.Sprintf("%s:%d", node.Host, node.Port)
	if _, _, err := net.SplitHostPort(addr); err != nil {
		return fmt.Errorf("无效的地址格式: %w", err)
	}

	if node.Timeout < 0 {
		return fmt.Errorf("超时时间不能为负数")
	}

	if node.PoolMinIdle < 0 {
		return fmt.Errorf("连接池最小空闲数不能为负数")
	}

	if node.PoolMaxActive < 0 {
		return fmt.Errorf("连接池最大活动数不能为负数")
	}

	if node.PoolMaxWait < 0 {
		return fmt.Errorf("连接池最大等待时间不能为负数")
	}

	return nil
}
