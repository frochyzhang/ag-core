package nettypoll

import (
	"ag-core/ag/ag_conf"
	"ag-core/ag/ag_ext/ip"
	mininetty "ag-core/ag/ag_netpoll"
	"context"
	"fmt"
	"log/slog"
	"time"
)

type Server struct {
	*mininetty.Server
	addr     string
	handlers []mininetty.ChannelHandler
	logger   *slog.Logger
}

type Option func(*Server)

func WithAddr(addr string) Option {
	return func(s *Server) {
		s.addr = addr
	}
}
func AppendHandler(ch mininetty.ChannelHandler) Option {
	return func(s *Server) {
		s.handlers = append(s.handlers, ch)
	}
}

func NewServer(logger *slog.Logger, opts ...Option) *Server {
	s := &Server{
		handlers: make([]mininetty.ChannelHandler, 0),
		logger:   logger,
	}

	for _, opt := range opts {
		opt(s)
	}

	initFunc := func(ch *mininetty.Channel) {
		pipeline := ch.Pipeline
		if pipeline != nil {
			for i, handler := range s.handlers {
				pipeline.AddLast(fmt.Sprintf("handler%d", i), handler)
			}
		}
	}

	server, err := mininetty.NewServer(s.addr, initFunc)
	if err != nil {
		panic(err)
	}
	s.Server = server
	return s
}

func NewNettyServerWithSuite(
	suite *MiniNettyOptionSuite,
	logger *slog.Logger,
) *Server {
	return NewServer(logger, suite.Options()...)
}

type MiniNettyOptionSuite struct {
	Opts []Option
}

func (s *MiniNettyOptionSuite) Options() []Option { return s.Opts }

type MiniNettySuiteBuilder struct {
	Binder        ag_conf.IBinder
	CustomOptions []Option
}

func (builder *MiniNettySuiteBuilder) BuildSuite() (*MiniNettyOptionSuite, error) {
	suite := &MiniNettyOptionSuite{
		Opts: make([]Option, 0),
	}

	suite.Opts = append(suite.Opts, builder.CustomOptions...)

	var conf MiniNettyServerProperties
	err := builder.Binder.Bind(&conf, miniNettyServerPropertiesPrefix)

	if err != nil {
		slog.Error("mininetty server config error", "error", err)
		return nil, err
	}

	host, port, err := findHostPort(conf)
	if err != nil {
		panic(err)
	}

	addr := fmt.Sprintf("%s:%d", host, port)

	slog.Info("mininetty", "host", addr)
	suite.Opts = append(suite.Opts, WithAddr(addr))

	return suite, nil
}

func (s *Server) Start(ctx context.Context) error {
	s.logger.Info("mininetty server start")
	s.Server.Start()
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("mininetty server shutdown")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s.Server.Shutdown()

	s.logger.Info("Shutting down mininetty server...")
	return nil
}

func findHostPort(conf MiniNettyServerProperties) (host string, port int, rerr error) {
	// 服务ip、端口配置
	host = conf.Host
	if host == "" {
		host = "0.0.0.0"
	}

	if !ip.IsHostAvailable(host) {
		return "", 0, fmt.Errorf("mininetty host unavailable: %s", host)
	}

	port = conf.Port
	if conf.AdaptivePort {
		slog.Info("mininetty server enable adaptive port")
		if port == 0 {
			port = DefaultHertzOriginPort
		}
		port, rerr = ip.GetAvailablePort(host, port)
		if rerr != nil {
			return
		}
	} else {
		if port == 0 {
			return host, port, fmt.Errorf("mininetty port invalid:%v", port)
		}
	}

	slog.Info(fmt.Sprintf("found mininetty host:%s, port:%d", host, port))
	return
}
