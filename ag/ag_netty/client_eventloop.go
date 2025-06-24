package ag_netty

import (
	"sync"
	"time"
)

// ClientEventLoop 客户端事件循环
type ClientEventLoop struct {
	taskQueue chan func()
	quit      chan struct{}
	wg        sync.WaitGroup
	initFunc  func(ch *Channel)
}

// NewClientEventLoop 创建客户端事件循环
func NewClientEventLoop(initFunc func(ch *Channel)) *ClientEventLoop {
	el := &ClientEventLoop{
		taskQueue: make(chan func(), 1024),
		quit:      make(chan struct{}),
		initFunc:  initFunc,
	}

	// 启动任务处理协程
	el.wg.Add(1)
	go el.runTaskLoop()

	return el
}

// Post 实现EventLooper接口
func (el *ClientEventLoop) Post(task func()) {
	select {
	case el.taskQueue <- task:
	case <-el.quit:
	}
}

func (el *ClientEventLoop) Schedule(delay time.Duration, task func()) {
	go func() {
		select {
		case <-time.After(delay):
			el.Post(task)
		case <-el.quit:
		}
	}()
}

func (el *ClientEventLoop) Shutdown() {
	close(el.quit)
	el.wg.Wait()
}

func (el *ClientEventLoop) IsShutdown() bool {
	select {
	case <-el.quit:
		return true
	default:
		return false
	}
}

// runTaskLoop 运行任务处理循环
func (el *ClientEventLoop) runTaskLoop() {
	defer el.wg.Done()

	for {
		select {
		case task := <-el.taskQueue:
			task()
		case <-el.quit:
			return
		}
	}
}
