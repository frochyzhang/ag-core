package fxs

import (
	"ag-core/ag/ag_conf"
	"ag-core/ag/ag_server"

	// "ag-core/ag/ag_server/kitex"

	agks "ag-core/ag/ag_kitex/server"

	"github.com/cloudwego/kitex/server"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"go.uber.org/fx"
)

// module
var FxKitexServerBaseModule = fx.Module(
	"fx_kitex_server_base",
	fx.Provide(
		/* === 1. 创建原生kitex server === */
		// kitex.NewHzwKCReporter,
		// kitex.NewKitexOriginalServer,

		// suite构建器，产出kitex server suite，由NewKitexServerWithSuit使用构建kitex server
		FxBuilderKitexServerSuite,

		// original kitex server
		agks.NewKitexServerWithSuit,

		// agkitex server middleware 入口，可按顺序执行agkitex server 自定义middleware
		FxBuildAgKitexServerMiddleware,

		/* === 2. 将原生kitex server 包装为ag server, 并注入到APP服务列表中 === */
		fx.Annotate(
			// kitex server 包装为ag server
			agks.NewServer,
			fx.As(new(ag_server.Server)), // 类型不匹配时，可以使用As指定接口类型
			fx.ResultTags(`group:"ag_servers"`),
		),
	),
	fx.Provide(
		// Ag 封装的middleware入口，可按顺序执行agkitex server 自定义middleware
		fx.Annotate(
			agks.AgRegistKitexServerMiddlewareOption,
			fx.ResultTags(`group:"kitex_server_options"`),
		),
	),
)

type FxInKitexServerParams struct {
	fx.In

	Env    ag_conf.IConfigurableEnvironment
	Binder ag_conf.IBinder

	CustOptions []*server.Option `group:"kitex_server_options",optional:"true"`

	NamingClient naming_client.INamingClient `optional:"true"`
}

func FxBuilderKitexServerSuite(params FxInKitexServerParams) (server.Suite, error) {
	build := &agks.KitexSuiteBuilder{
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

type FxAgKitexServerMiddlewareInParams struct {
	fx.In

	Middlewares []agks.IAgKitexServerMiddleware `group:"ag_kitex_server_middlewares",optional:"true"`
}

func FxBuildAgKitexServerMiddleware(p FxAgKitexServerMiddlewareInParams) *agks.AgKitexServerMiddleware {
	akm := &agks.AgKitexServerMiddleware{
		Middlewares: p.Middlewares,
	}
	return akm
}

// === kitex 服务端业务异常包装 middleware ===
var FxKitexAgServerBizErrorMiddlewareOption = fx.Provide(
	fx.Annotate(
		agks.NewAgBizErrorMiddlewareOption,
		// fx.As(new(*server.Option)),
		fx.ResultTags(`group:"kitex_server_options"`),
	),
)
