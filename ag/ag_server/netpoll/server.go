package main

import (
	mininetty "ag-core/ag/ag_netpoll"
	"fmt"
	"log/slog"
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

func NewNettyServer() *Server {
	//host := conf.GetProperty("hertz.server.host")
	//port, err := cast.ToIntE(conf.GetProperty("hertz.server.port"))
	//if err != nil {
	//	panic(err)
	//}

	//addr := fmt.Sprintf("%s:%d", host, port)

	return NewServer(
		&slog.Logger{},
		WithAddr(":8080"),
		AppendHandler(mininetty.NewLoggingHandler("server")),
		AppendHandler(&mininetty.EchoHandler{}),
	)
}

func main() {
	// 定义通道初始化函数
	//initFunc := func(ch *mininetty.Channel) {
	//	pipeline := ch.Pipeline
	//	pipeline.AddLast("logger", mininetty.NewLoggingHandler("server"))
	//	//pipeline.AddLast("decoder", &mininetty.ByteToMessageDecoder{})
	//	pipeline.AddLast("echo", &mininetty.EchoHandler{})
	//}

	server := NewNettyServer()

	// 创建服务器
	//server, err := mininetty.NewServer(":8080", initFunc)
	//if err != nil {
	//	panic(err)
	//}
	defer server.Server.Shutdown()
	//defer server.Shutdown()

	// 启动服务器
	server.Server.Start()
}
