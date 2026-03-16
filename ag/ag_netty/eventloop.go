package ag_netty

import (
	"sync"
	"time"
)

// EventLoop 事件循环 - 提供序列化任务执行能力
type EventLoop struct {
	taskQueue chan func()
	quit      chan struct{}
	wg        sync.WaitGroup
}

func NewEventLoop() *EventLoop {
	el := &EventLoop{
		taskQueue: make(chan func(), 1024),
		quit:      make(chan struct{}),
	}

	el.wg.Add(1)
	go el.runTaskLoop()

	return el
}

func (el *EventLoop) Post(task func()) {
	select {
	case el.taskQueue <- task:
	case <-el.quit:
	}
}

func (el *EventLoop) Schedule(delay time.Duration, task func()) {
	time.AfterFunc(delay, func() {
		el.Post(task)
	})
}

func (el *EventLoop) Shutdown() {
	close(el.quit)
	el.wg.Wait()
}

func (el *EventLoop) IsShutdown() bool {
	select {
	case <-el.quit:
		return true
	default:
		return false
	}
}

func (el *EventLoop) runTaskLoop() {
	defer el.wg.Done()

	for {
		select {
		case task := <-el.taskQueue:
			task()
		case <-el.quit:
			// drain remaining tasks before exit
			for {
				select {
				case task := <-el.taskQueue:
					task()
				default:
					return
				}
			}
		}
	}
}
