package fxs

import (
	"ag-core/ag/ag_conf"
	"ag-core/ag/ag_server"
	"ag-core/ag/ag_server/hertz"

	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"go.uber.org/fx"
)

// FxHertzOriginalServerBaseModule 创建裸HTTP服务
var FxHertzOriginalServerBaseModule = fx.Module("fx_hertz_original_server_base",
	fx.Provide(
		hertz.NewOriginalHertzServer,
	),

	fx.Provide(
		fx.Annotate(
			hertzServerWrapper,
			fx.ResultTags(`group:"ag_servers"`),
		),
	),
)

// FxHertzWithRegistryServerBaseModule 创建HTTP服务，并注册到注册中心
var FxHertzWithRegistryServerBaseModule = fx.Module("fx_hertz_with_registry_server",
	fx.Provide(
		FxBuilderHertzSuite,
		hertz.NewHertzServerWithSuit,
	),
	fx.Provide(
		fx.Annotate(
			hertzServerWrapper,
			fx.ResultTags(`group:"ag_servers"`),
		),
	),
)

func hertzServerWrapper(s *hertz.Server) ag_server.Server {
	return s
}

type FxHertzServerInParam struct {
	fx.In

	Env    ag_conf.IConfigurableEnvironment
	Binder ag_conf.IBinder

	CustOptions  []*config.Option            `group:"hertz_options" ,optional:"true"`
	NamingClient naming_client.INamingClient `optional:"true"`
}

func FxBuilderHertzSuite(params FxHertzServerInParam) (*hertz.HertzOptionSuite, error) {
	build := &hertz.HertzSuiteBuilder{
		Env:          params.Env,
		Binder:       params.Binder,
		NamingClient: params.NamingClient,
	}

	custOptions := make([]*config.Option, len(params.CustOptions))

	for _, opt := range params.CustOptions {
		custOptions = append(custOptions, opt)
	}

	return build.BuildSuite()
}
