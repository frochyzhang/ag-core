package ag_ext

import (
	"sync"
	"testing"
)

func TestCopyOnWriteMap(t *testing.T) {
	t.Run("PutAndGet", func(t *testing.T) {
		cowm := &CopyOnWriteMap[string, int]{}
		cowm.Put("a", 1)
		cowm.Put("b", 2)

		if v, ok := cowm.Get("a"); !ok || v != 1 {
			t.Errorf("Get(a) = %v, %v, want 1, true", v, ok)
		}
		if v, ok := cowm.Get("b"); !ok || v != 2 {
			t.Errorf("Get(b) = %v, %v, want 2, true", v, ok)
		}
		if _, ok := cowm.Get("c"); ok {
			t.Error("Get(c) should return false for non-existent key")
		}
	})

	t.Run("AsMap", func(t *testing.T) {
		cowm := &CopyOnWriteMap[string, int]{}
		cowm.Put("a", 1)
		cowm.Put("b", 2)

		m := cowm.AsMap()
		if len(m) != 2 {
			t.Errorf("AsMap() length = %d, want 2", len(m))
		}
		if m["a"] != 1 || m["b"] != 2 {
			t.Error("AsMap() returned incorrect values")
		}

		// 修改返回的map不应影响原数据
		m["a"] = 3
		if v, _ := cowm.Get("a"); v != 1 {
			t.Error("Modifying returned map affected original data")
		}
	})

	t.Run("ConcurrentAccess", func(t *testing.T) {
		cowm := &CopyOnWriteMap[int, int]{}
		var wg sync.WaitGroup

		// 并发写入
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				cowm.Put(i, i)
			}(i)
		}

		// 并发读取
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				cowm.Get(i)
			}(i)
		}

		wg.Wait()

		// 验证所有写入都成功
		for i := 0; i < 100; i++ {
			if v, ok := cowm.Get(i); !ok || v != i {
				t.Errorf("Get(%d) = %v, %v, want %d, true", i, v, ok, i)
			}
		}
	})
}

func TestCopyOnWriteSlice(t *testing.T) {
	t.Run("AddAndValue", func(t *testing.T) {
		cows := &CopyOnWriteSlice[int]{}
		cows.Add(1)
		cows.Add(2)

		s := cows.Value()
		if len(s) != 2 {
			t.Errorf("Value() length = %d, want 2", len(s))
		}
		if s[0] != 1 || s[1] != 2 {
			t.Error("Value() returned incorrect values")
		}
	})

	t.Run("AddIndex", func(t *testing.T) {
		cows := &CopyOnWriteSlice[string]{}
		cows.Add("a")
		cows.Add("c")

		// 中间插入
		cows.AddIndex(1, "b")
		// 头部插入
		cows.AddIndex(0, "0")
		// 尾部插入
		cows.AddIndex(100, "d")
		// 负索引插入
		cows.AddIndex(-1, "-1")

		s := cows.Value()
		expected := []string{"-1", "0", "a", "b", "c", "d"}
		if len(s) != len(expected) {
			t.Errorf("Value() length = %d, want %d", len(s), len(expected))
		}
		for i, v := range expected {
			if s[i] != v {
				t.Errorf("Value()[%d] = %s, want %s", i, s[i], v)
			}
		}
	})

	t.Run("IndexOf", func(t *testing.T) {
		cows := &CopyOnWriteSlice[int]{}
		cows.Add(1)
		cows.Add(2)
		cows.Add(3)

		if idx := cows.IndexOf(2); idx != 1 {
			t.Errorf("IndexOf(2) = %d, want 1", idx)
		}
		if idx := cows.IndexOf(4); idx != -1 {
			t.Errorf("IndexOf(4) = %d, want -1", idx)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		cows := &CopyOnWriteSlice[int]{}
		cows.Add(1)
		cows.Add(2)
		cows.Add(3)
		cows.Add(2)

		cows.Delete(2)
		s := cows.Value()
		if len(s) != 2 {
			t.Errorf("After Delete(2), length = %d, want 2", len(s))
		}
		if s[0] != 1 || s[1] != 3 {
			t.Error("Delete() removed wrong elements")
		}
	})

	t.Run("ConcurrentAccess", func(t *testing.T) {
		cows := &CopyOnWriteSlice[int]{}
		var wg sync.WaitGroup

		// 并发添加
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				cows.Add(i)
			}(i)
		}

		// 并发读取
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				cows.Value()
			}(i)
		}

		wg.Wait()

		// 验证所有添加都成功
		s := cows.Value()
		if len(s) != 100 {
			t.Errorf("After concurrent adds, length = %d, want 100", len(s))
		}
	})
}
