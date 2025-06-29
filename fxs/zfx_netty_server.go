package fxs

import (
	"github.com/frochyzhang/ag-core/ag/ag_conf"
	"github.com/frochyzhang/ag-core/ag/ag_netty"
	"github.com/frochyzhang/ag-core/ag/ag_netty/server"
	"github.com/frochyzhang/ag-core/ag/ag_server"
	"go.uber.org/fx"
)

// FxNettyServerBaseModule 创建裸HTTP服务
var FxNettyServerBaseModule = fx.Module("fx_mini_netty_server_base",

	fx.Provide(
		FxNewNettyServerSuite,
		fx.Annotate(
			server.NewNettyServerWithSuite,
			fx.As(new(ag_server.Server)),
			fx.ResultTags(`group:"ag_servers"`),
		),
	),
	fx.Provide(
		fx.Annotate(
			FxMnEchoOption,
			fx.ResultTags(`group:"ag_netty_server_options"`),
		),
	),
	fx.Provide(
		fx.Annotate(
			FxMnLoggerOption,
			fx.ResultTags(`group:"ag_netty_server_options"`),
		),
	),
)

type FxNettyServerInParam struct {
	fx.In

	Binder      ag_conf.IBinder
	CustOptions []server.Option `group:"ag_netty_server_options" ,optional:"true"`
}

func FxNewNettyServerSuite(params FxNettyServerInParam) (*server.NettyOptionSuite, error) {
	builder := &server.NettySuiteBuilder{
		Binder:        params.Binder,
		CustomOptions: params.CustOptions,
	}

	return builder.BuildSuite()
}

func FxMnLoggerOption() server.Option {
	return server.AppendHandler(ag_netty.NewLoggingHandler("fx_logger"))
}

func FxMnEchoOption() server.Option {
	return server.AppendHandler(&ag_netty.EchoHandler{})
}
