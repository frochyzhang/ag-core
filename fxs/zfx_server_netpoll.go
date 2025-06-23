package fxs

import (
	"ag-core/ag/ag_server"
	np "ag-core/ag/ag_server/netpoll"
	"context"
	"github.com/cloudwego/netpoll"

	"go.uber.org/fx"
)

// FxHertzOriginalServerBaseModule 创建裸HTTP服务
var FxNetpollServerBaseModule = fx.Module("fx_netpoll_server_base",
	fx.Provide(
		FxNetpollOptions,
	),

	fx.Provide(
		fx.Annotate(
			np.NewServer,
			fx.As(new(ag_server.Server)), // 类型不匹配时，可以使用As指定接口类型
			fx.ResultTags(`group:"ag_servers"`),
		),
	),
	fx.Invoke(func() {
		println("Starting Netpoll Server")
	}),
)

//type FxHertzServerInParam struct {
//	fx.In
//
//	Env    ag_conf.IConfigurableEnvironment
//	Binder ag_conf.IBinder
//
//	CustOptions  []*config.Option            `group:"hertz_options" ,optional:"true"`
//	NamingClient naming_client.INamingClient `optional:"true"`
//}

func FxNetpollOptions() np.Options {
	return &DefaultOptions{}
}

type DefaultOptions struct {
	np.Options
}

func (o *DefaultOptions) OnRequest(ctx context.Context, conn netpoll.Connection) error {
	return nil
}

func (o *DefaultOptions) OnConnect(ctx context.Context, conn netpoll.Connection) context.Context {
	println("OnConnect")
	return ctx
}

func (o *DefaultOptions) OnDisconnect(ctx context.Context, conn netpoll.Connection) {

}

func (o *DefaultOptions) OnPrepare(connection netpoll.Connection) context.Context {
	return context.Background()
}

func (o *DefaultOptions) OnClose(conn netpoll.Connection) error {
	return nil
}
