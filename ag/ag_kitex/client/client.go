package client

import (
	"github.com/frochyzhang/ag-core/ag/ag_ext"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/transport"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
)

type KitexClientSuite struct {
	opts []client.Option
}

func (s *KitexClientSuite) Options() []client.Option {
	return s.opts
}

type KitexSuiteBuilder struct {
	CustOptions []client.Option

	NamingClient naming_client.INamingClient
}

// func (builder *KitexSuiteBuilder) BuildSuite() (client.Suite, error) {
func (builder *KitexSuiteBuilder) BuildSuite() (*KitexClientSuite, error) {
	opts := make([]client.Option, 0)

	// 自定义的配置项
	opts = append(opts, builder.CustOptions...)

	// 指定传输协议，否则默认使用http1.1
	opts = append(opts, client.WithTransportProtocol(transport.GRPC))

	// 注册中心配置
	if builder.NamingClient != nil {
		opts = append(opts, client.WithResolver(ag_ext.NewNacosResolver(builder.NamingClient)))
	}

	// opts = append(opts, client.WithLoadBalancer())

	suite := &KitexClientSuite{
		opts: opts,
	}
	return suite, nil
}
