package fxs

import (
	"ag-core/ag/ag_log"
	"log/slog"

	"go.uber.org/fx"
)

var FxLogMode = fx.Module("ag_log",
	fx.Provide(ag_log.NewZapLog),
	fx.Provide(ag_log.NewZapHandler),
	fx.Provide(ag_log.NewZapSlog),

	// TODO 确定log加载
	fx.Invoke(func(l *slog.Logger) {
		l.Info("log init--------")
	}),
)
