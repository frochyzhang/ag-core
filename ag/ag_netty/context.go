package ag_netty

// HandlerContext 处理器上下文
type HandlerContext struct {
	name     string
	handler  ChannelHandler
	pipeline *Pipeline
	next     *HandlerContext
	prev     *HandlerContext
}

// newHandlerContext 创建处理器上下文
func newHandlerContext(name string, handler ChannelHandler, pipeline *Pipeline) *HandlerContext {
	return &HandlerContext{
		name:     name,
		handler:  handler,
		pipeline: pipeline,
	}
}

// Name 获取处理器名称
func (ctx *HandlerContext) Name() string {
	return ctx.name
}

// Pipeline 获取处理器流水线
func (ctx *HandlerContext) Pipeline() *Pipeline {
	return ctx.pipeline
}

// Channel 获取关联的通道
func (ctx *HandlerContext) Channel() *Channel {
	return ctx.pipeline.channel
}

// FireActive 触发激活事件
func (ctx *HandlerContext) FireActive() {
	if ctx.handler != nil {
		ctx.handler.HandleActive(ctx)
		ctx.next.FireActive()
	}
}

// FireInactive 触发失活事件
func (ctx *HandlerContext) FireInactive() {
	if ctx.handler != nil {
		ctx.handler.HandleInactive(ctx)
		ctx.next.FireInactive()
	}
}

// FireRead 触发读事件
func (ctx *HandlerContext) FireRead(data []byte) {
	if ctx.handler != nil {
		ctx.handler.HandleRead(ctx, data)
		ctx.next.FireRead(data)
	}
}

// FireWrite 触发写事件
func (ctx *HandlerContext) FireWrite(data []byte) {
	if ctx.handler != nil {
		ctx.handler.HandleWrite(ctx, data)
		ctx.prev.FireWrite(data)
	}
}

// FireError 触发错误事件
func (ctx *HandlerContext) FireError(err error) {
	if ctx.handler != nil {
		ctx.handler.HandleError(ctx, err)
		ctx.prev.FireError(err)
	}
}

// Write 写数据
func (ctx *HandlerContext) Write(data []byte) {
	ctx.Channel().Write(data)
}

// Close 关闭通道
func (ctx *HandlerContext) Close() {
	ctx.Channel().Close()
}
