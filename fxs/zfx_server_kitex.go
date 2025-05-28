package fxs

import (
	"ag-core/ag/ag_server"
	"ag-core/ag/ag_server/kitex"

	"go.uber.org/fx"
)

var FxKitexServerBaseModule = fx.Module(
	"fx_kitex_server_base",
	fx.Provide(
		kitex.NewHzwKCReporter,
		kitex.NewKitexOriginalServer,
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
