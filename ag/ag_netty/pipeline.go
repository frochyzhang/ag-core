package ag_netty

import "sync"

// Pipeline 处理器流水线
type Pipeline struct {
	head      *HandlerContext
	tail      *HandlerContext
	channel   *Channel
	handlerMu sync.RWMutex
}

// NewPipeline 创建处理器流水线
func NewPipeline(channel *Channel) *Pipeline {
	p := &Pipeline{
		channel: channel,
		head:    &HandlerContext{name: "HEAD"},
		tail:    &HandlerContext{name: "TAIL"},
	}

	p.head.next = p.tail
	p.tail.prev = p.head

	return p
}

// AddFirst 在头部添加处理器
func (p *Pipeline) AddFirst(name string, handler ChannelHandler) {
	p.handlerMu.Lock()
	defer p.handlerMu.Unlock()

	ctx := newHandlerContext(name, handler, p)

	next := p.head.next
	p.head.next = ctx
	ctx.prev = p.head
	ctx.next = next
	next.prev = ctx
}

// AddLast 在尾部添加处理器
func (p *Pipeline) AddLast(name string, handler ChannelHandler) {
	p.handlerMu.Lock()
	defer p.handlerMu.Unlock()

	ctx := newHandlerContext(name, handler, p)

	prev := p.tail.prev
	p.tail.prev = ctx
	ctx.next = p.tail
	ctx.prev = prev
	prev.next = ctx
}

// GetContext 获取指定名称的处理器上下文
func (p *Pipeline) GetContext(name string) *HandlerContext {
	p.handlerMu.RLock()
	defer p.handlerMu.RUnlock()

	for ctx := p.head.next; ctx != p.tail; ctx = ctx.next {
		if ctx.name == name {
			return ctx
		}
	}
	return nil
}

// FireActive 触发激活事件
func (p *Pipeline) FireActive() {
	p.head.next.FireActive()
}

// FireInactive 触发失活事件
func (p *Pipeline) FireInactive() {
	p.head.next.FireInactive()
}

// FireRead 触发读事件
func (p *Pipeline) FireRead(data []byte) {
	p.head.next.FireRead(data)
}

// FireWrite 触发写事件
func (p *Pipeline) FireWrite(data []byte) {
	p.tail.prev.FireWrite(data)
}

// FireError 触发错误事件
func (p *Pipeline) FireError(err error) {
	p.head.next.FireError(err)
}
