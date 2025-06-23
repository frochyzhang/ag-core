package ag_netpoll

import (
	"log/slog"
)

// EchoHandler 回显处理器
type EchoHandler struct{}

func (h *EchoHandler) HandleActive(ctx *HandlerContext) {
}

func (h *EchoHandler) HandleInactive(ctx *HandlerContext) {
	slog.Info("Connection closed: ", "remotePort", ctx.Channel().RemoteAddr())
}

func (h *EchoHandler) HandleRead(ctx *HandlerContext, data []byte) {
	// 回显接收到的数据
	ctx.Write(data)
}

func (h *EchoHandler) HandleWrite(ctx *HandlerContext, data []byte) {
	// 直接写入连接
	ctx.Channel().WriteDirect(data)
}

func (h *EchoHandler) HandleError(ctx *HandlerContext, err error) {
}
