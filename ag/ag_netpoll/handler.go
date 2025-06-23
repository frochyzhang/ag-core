package ag_netpoll

// ChannelHandler 通道处理器接口
type ChannelHandler interface {
	HandleActive(ctx *HandlerContext)
	HandleInactive(ctx *HandlerContext)
	HandleRead(ctx *HandlerContext, data []byte)
	HandleWrite(ctx *HandlerContext, data []byte)
	HandleError(ctx *HandlerContext, err error)
}
