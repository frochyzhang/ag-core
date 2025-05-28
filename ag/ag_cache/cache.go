package cache

import (
	"context"
	"errors"
	"sync"
)

var (
	// ErrNoSuchEngine 表示不支持的缓存引擎类型
	ErrNoSuchEngine = errors.New("no such cache engine")
	// constructors    = make(map[string]func(*Config) (ICache[any], error))
	once   sync.Once
	CacheM *CacheManager
	// CacheM = &CacheManager{
	// 	caches: make(map[string]CacheClearable),
	// }
)

type CacheManager struct {
	caches map[string]CacheClearable
}

func init() {
	// RegisterConstructor(CacheBigCacheEngineName, NewBigCache)
}

// fixEmpty 设置默认配置
func (cfg *Config) fixEmpty() {
	if cfg.MaxSizeInMB <= 0 {
		cfg.MaxSizeInMB = 100 // 默认100MB
	}

	if cfg.DefaultLoader == nil {
		cfg.DefaultLoader = func(context.Context, string) (any, error) {
			return nil, errors.New("no default loader")
		}
	}

}

// ICache 泛型缓存接口抽象
type ICache[T any] interface {
	// Get 获取key对应的值
	Get(key string) (T, bool)

	// GetWithCustLoader 获取key对应的值，如果不存在则调用loader加载
	GetWithCustLoader(ctx context.Context, key string, loader func(context.Context, string) (T, error)) (T, error)

	// GetWithLoader 获取key对应的值，如果不存在则调用默认加载器加载
	GetWithLoader(ctx context.Context, key string) (T, error)

	// Set 设置key-value对
	Set(key string, value T) error

	// SetWithExpire 设置带过期时间的key-value对
	SetWithExpire(key string, value T, ttl int64) error

	// Del 删除指定key
	Del(key string) error

	// Clear 清空所有缓存
	Clear() error

	// Stats 获取缓存统计信息
	Stats() Stats
}

// CacheClearable 缓存清理接口
type CacheClearable interface {
	Clear() error
}

type Config struct {
	SingleFlight  bool  // 是否使用singleflight
	MaxSizeInMB   int64 // 缓存最大大小(MB)
	DefaultLoader func(context.Context, string) (any, error)
}

// Option 配置选项函数类型
type Option func(cfg *Config)

// WithMaxSizeMB 设置缓存最大大小(MB)
func WithMaxSizeMB(maxSize int64) Option {
	return func(cfg *Config) {
		cfg.MaxSizeInMB = maxSize
	}
}

// WithDefaultLoader 设置默认加载器
func WithDefaultLoader(loader func(context.Context, string) (any, error)) Option {
	return func(cfg *Config) {
		cfg.DefaultLoader = loader
	}
}

// Stats 缓存统计信息
type Stats struct {
	// Hits is a number of successfully found keys
	Hits int64 `json:"hits"`
	// Misses is a number of not found keys
	Misses int64 `json:"misses"`
	// DelHits is a number of successfully deleted keys
	DelHits int64 `json:"delete_hits"`
	// DelMisses is a number of not deleted keys
	DelMisses int64 `json:"delete_misses"`
	// Collisions is a number of happened key-collisions
	Collisions int64 `json:"collisions"`
	// Len computes number of entries in cache
	Len int64 `json:"len"`
	// Cap amount of bytes store in the cache
	Cap int64 `json:"cap"`
}

// NewCache 创建泛型缓存实例
func NewCache[T any](engine string, opts ...Option) (ICache[T], error) {
	// 懒加载
	once.Do(func() {
		CacheM = &CacheManager{
			caches: make(map[string]CacheClearable),
		}
	})

	cfg := &Config{}
	for _, opt := range opts {
		opt(cfg)
	}

	cfg.fixEmpty()

	switch engine {
	case CacheBigCacheEngineName:
		return NewBigCache[T](cfg)
	default:
		return nil, ErrNoSuchEngine
	}

	// if f := GetConstructor(engine); f != nil {
	// 	cache, err := f(cfg)
	// 	if err != nil {
	// 		var zero ICache[T]
	// 		return zero, err
	// 	}
	// 	return cache.(ICache[T]), nil
	// }

	// var zero ICache[T]
	// return zero, ErrNoSuchEngine
}
