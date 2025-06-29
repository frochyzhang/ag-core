package fxs

import (
	"github.com/frochyzhang/ag-core/ag/ag_log"
	"log/slog"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

var FxLogMode = fx.Module("ag_log",
	fx.Provide(
		// 初始化zlog
		// ag_log.NewZapLog,
		ag_log.BindZlogProperties,
		ag_log.NewZapLogP,

		// 封装zap为slogHander
		ag_log.NewZapHandler,

		// 初始化slog
		ag_log.NewZapSlog,
	),

	// fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
	// 	return &fxevent.ZapLogger{
	// 		Logger: log,
	// 	}
	// }),
	fx.WithLogger(func(log *slog.Logger) fxevent.Logger {
		return &fxevent.SlogLogger{
			Logger: log,
		}
	}),

	// TODO 确定log加载
	fx.Invoke(func(l *slog.Logger) {
		l.Info("log init--------")
	}),
)
