package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/allegro/bigcache/v3"
	"golang.org/x/sync/singleflight"
)

var CacheBigCacheEngineName = "bigcache"

type CacheBigCache[T any] struct {
	cache        *bigcache.BigCache
	sg           *singleflight.Group
	defLoader    func(context.Context, string) (any, error)
	singleFlight bool
}

// func init() {
// 	RegisterEngine("bigcache", func(cfg *Config) (ICache[any], error) {
// 		return NewBigCache[any](cfg)
// 	})
// }

func NewBigCache[T any](cfg *Config) (ICache[T], error) {
	if cfg.MaxSizeInMB <= 0 {
		return nil, errors.New("MaxSizeInMB must be positive")
	}
	if cfg.MaxSizeInMB > 1<<31-1 {
		return nil, errors.New("MaxSizeInMB too large")
	}

	// TODO bigcache配置项调整取自Config，并增加相关option
	bcfg := bigcache.Config{
		Shards:             1024,
		LifeWindow:         0, // 永不过期
		CleanWindow:        0, // 不自动清理
		MaxEntriesInWindow: 1000 * 10 * 60,
		MaxEntrySize:       500,
		HardMaxCacheSize:   int(cfg.MaxSizeInMB),
		StatsEnabled:       true,
		Verbose:            true,
	}

	cache, initErr := bigcache.New(context.Background(), bcfg)
	if initErr != nil {
		return nil, initErr
	}

	cbc := &CacheBigCache[T]{
		cache:        cache,
		sg:           &singleflight.Group{},
		defLoader:    cfg.DefaultLoader,
		singleFlight: cfg.SingleFlight,
	}
	return cbc, nil
}

func (cbc *CacheBigCache[T]) Get(key string) (T, bool) {
	var zero T
	entry, err := cbc.cache.Get(key)
	if err != nil {
		if errors.Is(err, bigcache.ErrEntryNotFound) {
			return zero, false
		}
		return zero, false
	}

	var value T
	switch any(zero).(type) {
	case string:
		value = any(string(entry)).(T)
	case bool:
		// bool类型特殊处理
		b, err := strconv.ParseBool(string(entry))
		if err != nil {
			return zero, false
		}
		value = any(b).(T)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		// 数值类型直接类型断言
		value = any(entry).(T)
	default:
		// 复杂类型(包括指针)统一反序列化
		if err := json.Unmarshal(entry, &value); err != nil {
			return zero, false
		}
	}
	return value, true
}

func (cbc *CacheBigCache[T]) GetWithCustLoader(ctx context.Context, key string, loader func(context.Context, string) (T, error)) (T, error) {
	if val, ok := cbc.Get(key); ok {
		return val, nil
	}

	var val T

	if cbc.singleFlight {
		valt, err, _ := cbc.sg.Do(key, func() (any, error) {
			return loader(ctx, key)
		})
		// val, err := loader()
		if err != nil {
			var zero T
			return zero, err
		}
		valt2, ok := valt.(T)
		if !ok {
			var zero T
			return zero, errors.New("type assertion failed")
		}
		val = valt2
	} else {
		valt, err := loader(ctx, key)
		if err != nil {
			var zero T
			return zero, err
		}
		val = valt
	}

	if err := cbc.Set(key, val); err != nil {
		var zero T
		return zero, err
	}

	v, ok := cbc.Get(key)
	if !ok {
		var zero T
		return zero, errors.New("get failed")
	}
	return v, nil
}

func (cbc *CacheBigCache[T]) GetWithLoader(ctx context.Context, key string) (T, error) {
	if cbc.defLoader == nil {
		var zero T
		return zero, errors.New("no default loader")
	}
	return cbc.GetWithCustLoader(ctx, key, func(ctx2 context.Context, key2 string) (T, error) {
		v, err := cbc.defLoader(ctx2, key2)
		if err != nil {
			var zero T
			return zero, err
		}
		v2, ok := v.(T)
		if !ok {
			var zero T
			return zero, errors.New("type assertion failed")
		}
		return v2, nil
	})
}

func (cbc *CacheBigCache[T]) Set(key string, value T) error {
	var data []byte
	switch v := any(value).(type) {
	case string:
		data = []byte(v)
	case bool:
		// bool类型特殊处理
		data = []byte(strconv.FormatBool(v))
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		// 数值类型直接存储
		data = []byte(fmt.Sprintf("%v", v))
	default:

		/*
			// 处理指针类型
			if reflect.ValueOf(value).Kind() == reflect.Ptr {
				// 解引用指针再序列化
				if reflect.ValueOf(value).IsNil() {
					data = []byte("null")
				} else {
					var err error
					data, err = json.Marshal(reflect.ValueOf(value).Elem().Interface())
					if err != nil {
						return err
					}
				}
			} else {
				// 普通复杂类型序列化
				var err error
				data, err = json.Marshal(value)
				if err != nil {
					return err
				}
			}
		*/

		// 复杂类型(包括指针)统一序列化
		var err error
		data, err = json.Marshal(value)
		if err != nil {
			return err
		}
	}
	return cbc.cache.Set(key, data)
}

func (cbc *CacheBigCache[T]) SetWithExpire(key string, value T, ttl int64) error {
	// BigCache不支持单个key的TTL，统一使用全局LifeWindow
	return cbc.Set(key, value)
}
func (cbc *CacheBigCache[T]) Del(key string) error {
	return cbc.cache.Delete(key)
}

func (cbc *CacheBigCache[T]) Clear() error {
	return cbc.cache.Reset()
}

func (cbc *CacheBigCache[T]) Stats() Stats {
	cstats := cbc.cache.Stats()
	len := cbc.cache.Len()
	cap := cbc.cache.Capacity()

	return Stats{
		Hits:       cstats.Hits,
		Misses:     cstats.Misses,
		DelHits:    cstats.DelHits,
		DelMisses:  cstats.DelMisses,
		Collisions: cstats.Collisions,
		Len:        int64(len),
		Cap:        int64(cap),
	}
}
