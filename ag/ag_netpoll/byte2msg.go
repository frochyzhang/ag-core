package ag_netpoll

import (
	"log/slog"
)

// ByteToMessageDecoder 字节到消息解码器
type ByteToMessageDecoder struct{}

func (d *ByteToMessageDecoder) HandleActive(ctx *HandlerContext) {
}

func (d *ByteToMessageDecoder) HandleInactive(ctx *HandlerContext) {
	slog.Info("Connection closed: ", "remotePort", ctx.Channel().RemoteAddr())
}

func (d *ByteToMessageDecoder) HandleRead(ctx *HandlerContext, data []byte) {
	// 默认实现：直接传递字节数据
}

func (d *ByteToMessageDecoder) HandleWrite(ctx *HandlerContext, data []byte) {
}

func (d *ByteToMessageDecoder) HandleError(ctx *HandlerContext, err error) {
}
