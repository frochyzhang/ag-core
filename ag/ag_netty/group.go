package ag_netty

import "sync"

type EventLoopGroup struct {
	loops []*EventLoop
	next  int
	mu    sync.Mutex
}

func NewEventLoopGroup(size int) *EventLoopGroup {
	group := &EventLoopGroup{
		loops: make([]*EventLoop, size),
	}
	for i := 0; i < size; i++ {
		group.loops[i] = NewEventLoop()
	}
	return group
}

func (g *EventLoopGroup) Next() *EventLoop {
	g.mu.Lock()
	defer g.mu.Unlock()

	loop := g.loops[g.next]
	g.next = (g.next + 1) % len(g.loops)
	return loop
}

func (g *EventLoopGroup) Shutdown() {
	for _, loop := range g.loops {
		loop.Shutdown()
	}
}
