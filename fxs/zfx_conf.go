package fxs

import (
	"ag-core/ag/ag_conf"

	"go.uber.org/fx"
)

var FxAgConfModule = fx.Module("ag_conf",
	fx.Provide(
		fx.Annotate(
			ag_conf.NewStandardEnvironment,
			fx.As(new(ag_conf.IConfigurableEnvironment)),
		),
		fx.Annotate(
			ag_conf.NewConfigurationPropertiesBinder,
			fx.As(new(ag_conf.IBinder)),
		),
	),
)

var FxConfLocMode = fx.Module(
	"fx_conf_local",
	// LoadLocalConfig 构造使用了 embed.FS,目前需要应用main提前使用Supply等方式提供依赖
	fx.Invoke(ag_conf.LoadLocalConfig),
)
