package fxs

import (
	"github.com/frochyzhang/ag-core/ag/ag_nacos/config"
	"github.com/frochyzhang/ag-core/ag/ag_nacos/naming"

	"go.uber.org/fx"
)

var FxConfNacoMode = fx.Module(
	"fx_conf_nacos",
	fx.Provide(
		config.NewNacosConfigProperties,
		config.NewNacosConfigClient,
	),
)

var FxNacosNamingMode = fx.Module(
	"fx_nacos_naming",
	fx.Provide(
		naming.NewNacosNamingProperties,
		naming.NewNacosNamingClient,
	),
)

var FxEnableNacosRemoteConfigMode = fx.Module(
	"fx_nacos_remote_configenable",
	fx.Invoke(
		config.EnableNacosRemoteConfig,
	),
)
