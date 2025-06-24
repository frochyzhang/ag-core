package fxs

import (
	"ag-core/ag/ag_conf"
	"ag-core/ag/ag_netpoll"
	"ag-core/ag/ag_server"
	np "ag-core/ag/ag_server/netpoll"
	"go.uber.org/fx"
)

// FxMiniNettyServerBaseModule 创建裸HTTP服务
var FxMiniNettyServerBaseModule = fx.Module("fx_mini_netty_server_base",

	fx.Provide(
		FxNewMiniNettyServerSuite,
		fx.Annotate(
			np.NewNettyServerWithSuite,
			fx.As(new(ag_server.Server)),
			fx.ResultTags(`group:"ag_servers"`),
		),
	),
	fx.Provide(
		fx.Annotate(
			FxMnEchoOption,
			fx.ResultTags(`group:"ag_mn_custom_options"`),
		),
	),
	fx.Provide(
		fx.Annotate(
			FxMnLoggerOption,
			fx.ResultTags(`group:"ag_mn_custom_options"`),
		),
	),
)

type FxMiniNettyServerInParam struct {
	fx.In

	Binder      ag_conf.IBinder
	CustOptions []np.Option `group:"ag_mn_custom_options" ,optional:"true"`
}

func FxNewMiniNettyServerSuite(params FxMiniNettyServerInParam) (*np.MiniNettyOptionSuite, error) {
	builder := &np.MiniNettySuiteBuilder{
		Binder:        params.Binder,
		CustomOptions: params.CustOptions,
	}

	return builder.BuildSuite()
}

func FxMnLoggerOption() np.Option {
	return np.AppendHandler(ag_netpoll.NewLoggingHandler("fx_logger"))
}

func FxMnEchoOption() np.Option {
	return np.AppendHandler(&ag_netpoll.EchoHandler{})
}
