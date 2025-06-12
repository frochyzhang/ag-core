package ag_log

import (
	"context"
	"log/slog"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ZapHandler 是一个实现了 slog.Handler 接口的结构体
type ZapHandler struct {
	zaplogger *zap.Logger
}

// NewZapHandler 创建一个新的 ZapHandler 实例
func NewZapHandler(logger *zap.Logger) *ZapHandler {
	// CallerSkip调整为3，获取真正调用处的位置
	log := logger.WithOptions(zap.AddCallerSkip(3))
	return &ZapHandler{
		zaplogger: log,
	}
}

// / NewZapSlog 创建zap封装的slog
func NewZapSlog(zapHandler *ZapHandler) *slog.Logger {
	slogger := slog.New(zapHandler)
	slogger.Info("slog init", "hander", zapHandler)
	slog.SetDefault(slogger) // slog设置全局实现
	return slogger
}

// Enabled 检查日志级别是否启用
func (h *ZapHandler) Enabled(ctx context.Context, level slog.Level) bool {
	zapLevel := LevelSlog2Zap(level)
	return h.zaplogger.Core().Enabled(zapLevel)
}

// Handle 处理日志记录
func (h *ZapHandler) Handle(ctx context.Context, r slog.Record) error {
	fields := make([]zap.Field, 0, r.NumAttrs())
	// TEST
	//	fields = append(fields, zap.Any("TEST", "slog4zap"))

	r.Attrs(func(attr slog.Attr) bool {
		fields = append(fields, zap.Any(attr.Key, attr.Value.Any()))
		return true
	})

	switch {
	case r.Level >= slog.LevelError:
		h.zaplogger.Error(r.Message, fields...)
	case r.Level >= slog.LevelWarn:
		h.zaplogger.Warn(r.Message, fields...)
	case r.Level >= slog.LevelInfo:
		h.zaplogger.Info(r.Message, fields...)
	case r.Level >= slog.LevelDebug:
		h.zaplogger.Debug(r.Message, fields...)
	}
	return nil
}

// WithAttrs 返回一个新的带有额外属性的 Handler
func (h *ZapHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// TODO 待优化
	h.zaplogger = h.zaplogger.With(zap.Any(attrs[0].Key, attrs[0].Value.Any()))
	return h
}

// WithGroup 返回一个新的带有日志组的 Handler
func (h *ZapHandler) WithGroup(name string) slog.Handler {
	// TODO 待实现
	return h
}

// LevelSlog2Zap 将 slog.Level 转换为 zapcore.Level
func LevelSlog2Zap(level slog.Level) zapcore.Level {
	switch {
	case level >= slog.LevelError:
		return zapcore.ErrorLevel
	case level >= slog.LevelWarn:
		return zapcore.WarnLevel
	case level >= slog.LevelInfo:
		return zapcore.InfoLevel
	case level >= slog.LevelDebug:
		return zapcore.DebugLevel
	default:
		return zapcore.InfoLevel
	}
}
