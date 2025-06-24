package fxs

import (
	"ag-core/ag/ag_conf"
	"ag-core/ag/ag_netty"
	"ag-core/ag/ag_netty/client"
	"errors"
	"go.uber.org/fx"
)

// FxNettyClientBaseModule 创建裸netty client
var FxNettyClientBaseModule = fx.Module("fx_netty_client_base",

	fx.Provide(
		FxNewNettyClientWithSuite,
	),
	fx.Provide(
		fx.Annotate(
			FxClientConnectorOption,
			fx.ResultTags(`group:"ag_netty_client_options"`),
		),
	),
	fx.Provide(
		fx.Annotate(
			FxClientLoggerOption,
			fx.ResultTags(`group:"ag_netty_client_options"`),
		),
	),
	fx.Provide(
		fx.Annotate(
			FxClientEchoOption,
			fx.ResultTags(`group:"ag_netty_client_options"`),
		),
	),
	fx.Invoke(FxNewNettyClientWithSuite),
)

type FxNettyClientInParam struct {
	fx.In

	Env ag_conf.IConfigurableEnvironment

	CustomOptions []client.Option `group:"ag_netty_client_options" ,optional:"true"`
}

func FxNewNettyClientWithSuite(params FxNettyClientInParam) (*client.NettyOptionSuite, error) {
	remoteAddr := params.Env.GetProperty("netty.remote.addr")
	if remoteAddr == "" {
		return nil, errors.New("netty.remote.addr is required")
	}
	opts := params.CustomOptions
	opts = append(opts, client.WithAddr(remoteAddr))
	return &client.NettyOptionSuite{
		Opts: opts,
	}, nil
}

func FxClientConnectorOption() client.Option {
	return client.AppendHandler(&ag_netty.ConnectorHandler{})
}

func FxClientLoggerOption() client.Option {
	return client.AppendHandler(ag_netty.NewLoggingHandler("fx_logger"))
}

func FxClientEchoOption() client.Option {
	return client.AppendHandler(&client.EchoHandler{EchoHandler: &ag_netty.EchoHandler{}})
}
