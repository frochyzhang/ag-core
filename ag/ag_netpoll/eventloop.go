package ag_netpoll

import (
	"context"
	"errors"
	"github.com/cloudwego/netpoll"
	"net"
	"sync"
	"time"
)

// EventLoop 事件循环
type EventLoop struct {
	loop      netpoll.EventLoop
	taskQueue chan func()
	quit      chan struct{}
	wg        sync.WaitGroup
	connMap   sync.Map // 存储连接的映射
	initFunc  func(ch *Channel)
}

// NewEventLoop 创建新事件循环
func NewEventLoop(initFunc func(ch *Channel)) (*EventLoop, error) {
	el := &EventLoop{
		taskQueue: make(chan func(), 1024),
		quit:      make(chan struct{}),
		initFunc:  initFunc,
	}

	// 使用闭包捕获 EventLoop 实例
	loop, err := netpoll.NewEventLoop(
		el.handleRequest,
		netpoll.WithOnPrepare(el.handlePrepare),
		netpoll.WithOnDisconnect(el.handleDisconnect),
		netpoll.WithIdleTimeout(time.Hour),
	)
	if err != nil {
		return nil, err
	}

	el.loop = loop

	// 启动任务处理协程
	el.wg.Add(1)
	go el.runTaskLoop()

	return el, nil
}

// handlePrepare 处理新连接准备
func (el *EventLoop) handlePrepare(conn netpoll.Connection) context.Context {
	// 创建新通道
	channel := NewChannel(conn, el)
	el.connMap.Store(conn, channel)

	if el.initFunc != nil {
		el.initFunc(channel)
	}
	// 触发激活事件
	channel.Pipeline.FireActive()

	return context.WithValue(context.Background(), "channel", channel)
}

func (el *EventLoop) handleDisconnect(ctx context.Context, conn netpoll.Connection) {
	c, ok := el.connMap.LoadAndDelete(conn)
	if !ok {
		return
	}
	context.WithValue(ctx, "channel", nil)
	c.(*Channel).Close()
}

// handleRequest 处理请求
func (el *EventLoop) handleRequest(ctx context.Context, conn netpoll.Connection) error {
	// 从上下文中获取通道
	channel, ok := ctx.Value("channel").(*Channel)
	if !ok || channel == nil {
		return errors.New("channel not found in context")
	}

	// 读取数据
	reader := conn.Reader()
	n := reader.Len()
	if n == 0 {
		return nil
	}

	data, err := reader.ReadBinary(n)
	if err != nil {
		channel.Pipeline.FireError(err)
		return err
	}

	// 触发读事件
	channel.Pipeline.FireRead(data)
	return nil
}

// runTaskLoop 运行任务处理循环
func (el *EventLoop) runTaskLoop() {
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

// Post 投递任务到事件循环
func (el *EventLoop) Post(task func()) {
	select {
	case el.taskQueue <- task:
	case <-el.quit:
	}
}

// Schedule 调度延迟任务
func (el *EventLoop) Schedule(delay time.Duration, task func()) {
	time.AfterFunc(delay, func() {
		el.Post(task)
	})
}

// Run 运行事件循环
func (el *EventLoop) Run(listener net.Listener) error {
	return el.loop.Serve(listener)
}

// Shutdown 关闭事件循环
func (el *EventLoop) Shutdown() {
	close(el.quit)
	el.wg.Wait()
	el.loop.Shutdown(context.Background())

	// 关闭所有连接
	el.connMap.Range(func(key, value interface{}) bool {
		if ch, ok := value.(*Channel); ok {
			ch.Close()
		}
		return true
	})
}
