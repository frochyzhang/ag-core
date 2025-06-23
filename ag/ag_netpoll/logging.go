package ag_netpoll

import (
	"log/slog"
)

// LoggingHandler 日志处理器
type LoggingHandler struct {
	name string
}

func NewLoggingHandler(name string) *LoggingHandler {
	return &LoggingHandler{name: name}
}

func (h *LoggingHandler) HandleActive(ctx *HandlerContext) {
	slog.Info("Connection active: ", "remotePort", ctx.Channel().RemoteAddr())
}

func (h *LoggingHandler) HandleInactive(ctx *HandlerContext) {
	slog.Info("Connection closed: ", "remotePort", ctx.Channel().RemoteAddr())
}

func (h *LoggingHandler) HandleRead(ctx *HandlerContext, data []byte) {
	slog.Info(" Read bytes!", "length", len(data))
}

func (h *LoggingHandler) HandleWrite(ctx *HandlerContext, data []byte) {
	slog.Info(" Write bytes!", "length", len(data))
}

func (h *LoggingHandler) HandleError(ctx *HandlerContext, err error) {
	slog.Error("Connection error", "error", err)
}
