package ag_netpoll

import "sync"

// EventLoopGroup 事件循环组
type EventLoopGroup struct {
	loops []*EventLoop
	next  int
	mu    sync.Mutex
}

// NewEventLoopGroup 创建事件循环组
func NewEventLoopGroup(size int, initFunc func(ch *Channel)) (*EventLoopGroup, error) {
	group := &EventLoopGroup{
		loops: make([]*EventLoop, size),
	}

	for i := 0; i < size; i++ {
		loop, err := NewEventLoop(initFunc)
		if err != nil {
			// 清理已创建的循环
			for j := 0; j < i; j++ {
				group.loops[j].Shutdown()
			}
			return nil, err
		}
		group.loops[i] = loop
	}

	return group, nil
}

// Next 获取下一个事件循环
func (g *EventLoopGroup) Next() *EventLoop {
	g.mu.Lock()
	defer g.mu.Unlock()

	loop := g.loops[g.next]
	g.next = (g.next + 1) % len(g.loops)
	return loop
}

// Shutdown 关闭事件循环组
func (g *EventLoopGroup) Shutdown() {
	for _, loop := range g.loops {
		loop.Shutdown()
	}
}
