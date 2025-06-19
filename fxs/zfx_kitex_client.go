package fxs

import (
	agkc "ag-core/ag/ag_kitex/client"

	"github.com/cloudwego/kitex/client"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"go.uber.org/fx"
)

var FxKitexClientBaseModule = fx.Module(
	"fx_kitex_Client_base",
	fx.Provide(
		FxBuilderKitexClientSuite,
	),
)

var FxKitexAgClientBizErrorMiddlewareOption = fx.Provide(
	fx.Annotate(
		agkc.NewAgBizErrorMiddlewareOption,
		fx.ResultTags(`group:"kitex_client_options"`),
	),
)

type FxInKitexClientParams struct {
	fx.In

	CustOptions []*client.Option `group:"kitex_client_options",optional:"true"`

	NamingClient naming_client.INamingClient `optional:"true"`
}

// func FxBuilderKitexClientSuite(params FxInKitexServerParams) (server.Suite, error) {
func FxBuilderKitexClientSuite(params FxInKitexClientParams) (*agkc.KitexClientSuite, error) {
	build := &agkc.KitexSuiteBuilder{
		NamingClient: params.NamingClient,
	}
	// CustOptions:  params.CustOptions,
	custOpt := make([]client.Option, 0)
	for _, opt := range params.CustOptions {
		custOpt = append(custOpt, *opt)
	}
	build.CustOptions = custOpt

	return build.BuildSuite()
}
