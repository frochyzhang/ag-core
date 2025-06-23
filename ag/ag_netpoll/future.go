package ag_netpoll

import (
	"errors"
	"sync"
	"time"
)

// Future 异步操作结果
type Future struct {
	done   chan struct{}
	result interface{}
	mu     sync.Mutex
}

// NewFuture 创建新Future
func NewFuture() *Future {
	return &Future{done: make(chan struct{})}
}

// Complete 完成Future
func (f *Future) Complete(result interface{}) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.done == nil {
		return
	}

	f.result = result
	close(f.done)
}

// Get 获取结果（阻塞）
func (f *Future) Get() interface{} {
	<-f.done
	return f.result
}

// GetWithTimeout 带超时获取结果
func (f *Future) GetWithTimeout(timeout time.Duration) (interface{}, error) {
	select {
	case <-f.done:
		return f.result, nil
	case <-time.After(timeout):
		return nil, errors.New("future timeout")
	}
}

// IsDone 判断是否完成
func (f *Future) IsDone() bool {
	select {
	case <-f.done:
		return true
	default:
		return false
	}
}
