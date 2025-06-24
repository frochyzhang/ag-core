package ag_netty

import "log/slog"

// ConnectorHandler 连接处理器
type ConnectorHandler struct{}

func (h *ConnectorHandler) HandleActive(ctx *HandlerContext) {
	slog.Info("Connection established", "remote", ctx.Channel().RemoteAddr())
}

func (h *ConnectorHandler) HandleInactive(ctx *HandlerContext) {
	slog.Info("Connection closed", "remote", ctx.Channel().RemoteAddr())
}

func (h *ConnectorHandler) HandleRead(ctx *HandlerContext, data []byte) {
	// 默认不处理，由用户处理器处理
}

func (h *ConnectorHandler) HandleWrite(ctx *HandlerContext, data []byte) {
	// 直接写入连接
	//ctx.Channel().WriteDirect(data)
}

func (h *ConnectorHandler) HandleError(ctx *HandlerContext, err error) {
	slog.Error("Connection error", "error", err, "remote", ctx.Channel().RemoteAddr())
	ctx.Close()
}
