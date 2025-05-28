package kitex

import (
	"ag-core/ag/ag_conf"
	"context"
	"fmt"
	"log/slog"
	"net"

	"github.com/cloudwego/kitex/pkg/remote/connpool"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/pkg/stats"
	"github.com/cloudwego/kitex/pkg/transmeta"
	"github.com/cloudwego/kitex/server"
	"github.com/kitex-contrib/registry-nacos/registry"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/spf13/cast"
)

type Server struct {
	ksvr   server.Server
	logger *slog.Logger
}

// NewServer (engine *gin.Engine, logger *log.Logger, opts ...Option) *Server {
func NewServer(
	svr server.Server,
	logger *slog.Logger,
	creporter *HzwKCReporter,
) *Server {
	s := &Server{
		ksvr:   svr,
		logger: logger,
	}
	connpool.SetReporter(creporter) // 注册自定义连接池监控，TODO是否有用？
	return s
}

func (s *Server) Start(ctx context.Context) error {
	sinfos := s.ksvr.GetServiceInfos()
	// 格式化打印服务信息
	for sname, sinfo := range sinfos {
		s.logger.Info("Kitex服务信息",
			"服务名称", sname,
			"方法数量", len(sinfo.Methods),
			"方法列表", sinfo.Methods,
			"额外信息", sinfo.Extra,
		)
	}

	s.logger.Info("Kitex服务启动",
		"总服务数", len(sinfos),
	)
	return s.ksvr.Run()
}
func (s *Server) Stop(ctx context.Context) error {
	return s.ksvr.Stop()
}

// HzwKCReporter 自定义连接池监控
type HzwKCReporter struct{}

func NewHzwKCReporter() *HzwKCReporter {
	return &HzwKCReporter{}
}

func (r *HzwKCReporter) ConnSucceed(poolType connpool.ConnectionPoolType, serviceName string, addr net.Addr) {
	fmt.Printf("ConnSucceed poolType:%d, serviceName:%s, addr:%v\n", poolType, serviceName, addr)
}
func (r *HzwKCReporter) ConnFailed(poolType connpool.ConnectionPoolType, serviceName string, addr net.Addr) {
	fmt.Printf("ConnFailed poolType:%d, serviceName:%s, addr:%v\n", poolType, serviceName, addr)
}
func (r *HzwKCReporter) ReuseSucceed(poolType connpool.ConnectionPoolType, serviceName string, addr net.Addr) {
	fmt.Printf("ReuseSucceed poolType:%d, serviceName:%s, addr:%v\n", poolType, serviceName, addr)
}

// NewKitexOriginalServer 创建原始的kitex服务实例
func NewKitexOriginalServer(
	logger *slog.Logger,
	env ag_conf.IConfigurableEnvironment,
	client naming_client.INamingClient,
) server.Server {
	host := env.GetProperty("kitex.host")
	port, err := cast.ToIntE(env.GetProperty("kitex.port"))
	if err != nil {
		panic(err)
	}
	endpoint := env.GetProperty("kitex.endpoint")

	kitexHostStr := fmt.Sprintf("%s:%d", host, port)
	logger.Info("kitex", "host", kitexHostStr)

	addr, _ := net.ResolveTCPAddr("tcp", kitexHostStr)
	/*
		svr := server.NewServer(
			server.WithServiceAddr(addr),
			// 配合 SkipDecoder https://www.cloudwego.io/zh/docs/kitex/tutorials/code-gen/skip_decoder/
			server.WithPayloadCodec(thrift.NewThriftCodecWithConfig(thrift.FastWrite|thrift.FastRead|thrift.EnableSkipDecoder)),
			//server.WithMuxTransport(),                               // 服务端开启多路复用；Server 开启连接多路复用对 Client 没有限制，可以接受短连接、长连接池、连接多路复用的请求
			//server.WithMetaHandler(transmeta.ServerHTTP2Handler),    // 指定基于HTTP2 协议header的信息透传
			server.WithMetaHandler(transmeta.ServerTTHeaderHandler), // 指定基于TTHeader 协议header的信息透传
			server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: endpoint}),
			server.WithRegistry(registry.NewNacosRegistry(client)),
		)
	*/

	options := []server.Option{
		server.WithServiceAddr(addr),
		server.WithMetaHandler(transmeta.ServerTTHeaderHandler),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: endpoint}),
		// server.WithStatsLevel(stats.LevelDetailed),
		server.WithStatsLevel(stats.LevelDisabled),
	}
	if client != nil {
		slog.Info("kitex server enable nacos naming")
		options = append(options, server.WithRegistry(registry.NewNacosRegistry(client)))
	}
	svr := server.NewServer(options...)

	return svr
}
