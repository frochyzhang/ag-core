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
	fx.Provide(
		ag_conf.LoadLocalConfigToState, // 在provide阶段解析初始化本地配置，并返回一个本地初始化完成的标志，方便其他要依赖本地配置的组件控制初始化顺序
	),
	// fx.Invoke(
	// 	ag_conf.LoadLocalConfig,
	// ),
)
