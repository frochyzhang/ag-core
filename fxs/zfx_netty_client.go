package fxs

import (
	"github.com/frochyzhang/ag-core/ag/ag_conf"
	"github.com/frochyzhang/ag-core/ag/ag_netty"
	"github.com/frochyzhang/ag-core/ag/ag_netty/client"
	"go.uber.org/fx"
	"log/slog"
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
			//fx.ResultTags(`group:"ag_netty_client_options"`),
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

	Env    ag_conf.IConfigurableEnvironment
	Binder ag_conf.IBinder

	CustomOptions []client.Option `group:"ag_netty_client_options" ,optional:"true"`
}

func FxNewNettyClientWithSuite(params FxNettyClientInParam) (*client.NettyOptionSuite, error) {
	var clientProps client.NettyClientProperties
	err := params.Binder.Bind(&clientProps, client.NettyClientPropertiesPrefix)
	if err != nil {
		slog.Error("ag_netty client config error", "error", err)
		return nil, err
	}

	opts := params.CustomOptions
	opts = append(opts, client.WithProps(clientProps))
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
