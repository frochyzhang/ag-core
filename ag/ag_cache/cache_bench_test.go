package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"runtime"
	"testing"
	"time"
)

func TestCacheHello(t *testing.T) {
	// 创建字符串类型缓存
	strCache, err := NewCache[string]("bigcache", WithMaxSizeMB(500))
	if err != nil {
		t.Fatal(err)
	}
	strCache.Set("name", "张三")

	// 创建结构体类型缓存
	type Hzw struct {
		Name string
		Age  int
	}
	userCache, err := NewCache[Hzw]("bigcache")
	userCache.Set("u1", Hzw{"李四", 30})

	user, ok := userCache.Get("u1")
	if !ok {
		t.Fatal("get u1 failed")
	}
	slog.Info("user", "user", user)

	// bool类型缓存
	boolCache, err := NewCache[bool]("bigcache", WithMaxSizeMB(2))
	if err != nil {
		t.Fatal(err)
	}
	boolCache.Set("b1", true)
	b, ok := boolCache.Get("b1")
	slog.Info("b1", "b1", b, "ok", ok)
	stats := boolCache.Stats()
	slog.Info("stats", "stats", stats)

	// any类型缓存
	anyCache, err := NewCache[any]("bigcache")
	if err != nil {
		t.Fatal(err)
	}
	anyCache.Set("a1", "hello")
	anyCache.Set("a2", 123)
	anyCache.Set("a3", true)
	anyCache.Set("a4", 123.456)

	anyCache.Get("a1")
	anyCache.Get("a2")
	anyCache.Get("a3")
	anyCache.Get("a4")
	anyCache.Get("a5")
	stats = anyCache.Stats()
	slog.Info("stats", "stats", stats)

	time.Sleep(time.Second * 1)
}

func TestCacheWithLoader(t *testing.T) {
	// 创建字符串类型缓存
	cache, err := NewCache[string](CacheBigCacheEngineName, WithMaxSizeMB(2), WithDefaultLoader(func(ctx context.Context, key string) (any, error) {
		return "hello" + key, nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	cache.GetWithLoader(ctx, "hhh")
	cache.GetWithLoader(ctx, "zzz")
	cache.GetWithLoader(ctx, "www")

	for i := 0; i < 1000; i++ {
		cache.GetWithLoader(ctx, "zzz")
	}

	stats := cache.Stats()
	slog.Info("stats", "stats", stats)

	// 创建struct 指针类型缓存
	cache2, err := NewCache[*hzw](
		CacheBigCacheEngineName,
		WithMaxSizeMB(2),
		WithDefaultLoader(func(ctx context.Context, key string) (any, error) {
			_hzw := hzw{Id: 123}
			// 打印_hzw的内存地址
			fmt.Printf("loader _hzw: %p\n", &_hzw)
			return &_hzw, nil
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	v, err := cache2.GetWithLoader(ctx, "hhh")
	fmt.Printf("_hzw: %p  %v\n", v, v)
	v, err = cache2.GetWithLoader(ctx, "hhh")
	fmt.Printf("_hzw: %p  %v\n", v, v)
	v = nil
	// 触发GC
	runtime.GC()
	// time.Sleep(time.Second)
	v, err = cache2.GetWithLoader(ctx, "hhh")
	fmt.Printf("_hzw: %p  %v\n", v, v)
	cache2.Clear()
	v, err = cache2.GetWithLoader(ctx, "hhh")
	fmt.Printf("_hzw: %p  %v\n", v, v)

	// 创建string 指针类型缓存, 测试bigcache的内存回收,其处理逻辑和hzw指针一样
	cache3, _ := NewCache[*string](
		CacheBigCacheEngineName,
		WithMaxSizeMB(2),
		WithDefaultLoader(func(ctx context.Context, key string) (any, error) {
			str := fmt.Sprintf("str 4 %s", key)
			return &str, nil
		}),
	)

	v2, err := cache3.GetWithLoader(ctx, "zzz")
	fmt.Printf("*string: %p %v: %v\n", v2, v2, *v2)
	v2, err = cache3.GetWithLoader(ctx, "zzz")
	fmt.Printf("*string: %p %v: %v\n", v2, v2, *v2)
	v2 = nil
	runtime.GC()
	v2, err = cache3.GetWithLoader(ctx, "zzz")
	fmt.Printf("*string: %p %v: %v\n", v2, v2, *v2)

}

const maxEntrySize = 256
const maxEntryCount = 10000

type hzw struct {
	Id int `json:"id"`
}

func value() []byte {
	return make([]byte, 100)
}
func parallelKey(threadID int, counter int) string {
	return fmt.Sprintf("key-%04d-%06d", threadID, counter)
}

type constructor[T any] interface {
	Get(int) T
	Parse([]byte) (T, error)
	ToBytes(T) ([]byte, error)
}

type byteConstructor []byte

func (bc byteConstructor) Get(n int) []byte {
	return value()
}

func (bc byteConstructor) Parse(data []byte) ([]byte, error) {
	return data, nil
}

func (bc byteConstructor) ToBytes(v []byte) ([]byte, error) {
	return v, nil
}

type structConstructor struct {
}

func (sc structConstructor) Get(n int) hzw {
	return hzw{Id: n}
}

func (sc structConstructor) Parse(data []byte) (hzw, error) {
	var s hzw
	err := json.Unmarshal(data, &s)
	return s, err
}

func (sc structConstructor) ToBytes(v hzw) ([]byte, error) {
	return json.Marshal(v)
}

func BenchmarkCacheByBigCacheForStruct(b *testing.B) {
	CacheParallel[hzw](CacheBigCacheEngineName, structConstructor{}, b)
}

func CacheParallel[T any](engine string, cs constructor[T], b *testing.B) {

	cache, err := NewCache[T](engine)
	if err != nil {
		b.Fatal(err)
	}

	b.RunParallel(func(pb *testing.PB) {
		thread := rand.Intn(1000)
		for pb.Next() {
			id := rand.Intn(maxEntryCount)
			// data, _ := cs.ToBytes(cs.Get(id))
			// cache.Set(parallelKey(thread, id), data)
			cache.Set(parallelKey(thread, id), cs.Get(id))
		}
	})

	// b.RunParallel(func(pb *testing.PB) {
	// 	thread := rand.Intn(1000)
	// 	for pb.Next() {
	// 		id := rand.Intn(maxEntryCount)
	// 		cache.Get(parallelKey(thread, id))
	// 	}
	// })
	// stats := cache.Stats()
	// slog.Info("stats", "stats", stats)

}
