package fxs

import (
	"ag-core/ag/ag_conf"
	"ag-core/ag/ag_server"
	"ag-core/ag/ag_server/kitex"

	"github.com/cloudwego/kitex/server"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"go.uber.org/fx"
)

var FxKitexServerBaseModule = fx.Module(
	"fx_kitex_server_base",
	fx.Provide(
		// kitex.NewHzwKCReporter,
		// kitex.NewKitexOriginalServer,
		// suite构建器
		FxBuilderKitexSuite,
		// original kitex server
		kitex.NewKitexServerWithSuit,
		// ag Server
		kitex.NewServer,
	),
	fx.Provide(
		fx.Annotate(
			kitexserverWrapper,
			fx.As(new(ag_server.Server)), // 类型不匹配时，可以使用As指定接口类型
			fx.ResultTags(`group:"ag_servers"`),
		),
	),
)

func kitexserverWrapper(s *kitex.Server) ag_server.Server {
	return s
}

type FxKitexServerParams struct {
	fx.In

	Env    ag_conf.IConfigurableEnvironment
	Binder ag_conf.IBinder

	CustOptions []*server.Option `group:"kitex_options",optional:"true"`

	NamingClient naming_client.INamingClient `optional:"true"`
}

func FxBuilderKitexSuite(params FxKitexServerParams) (server.Suite, error) {
	build := &kitex.KitexSuiteBuilder{
		Env:          params.Env,
		Binder:       params.Binder,
		NamingClient: params.NamingClient,
	}
	// CustOptions:  params.CustOptions,
	custOpt := make([]server.Option, 0)
	for _, opt := range params.CustOptions {
		custOpt = append(custOpt, *opt)
	}
	build.CustOptions = custOpt

	return build.BuildSuite()
}
