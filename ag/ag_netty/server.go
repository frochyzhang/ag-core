package ag_netty

import (
	"log/slog"
	"net"
)

// Server 服务器
type Server struct {
	group    *EventLoopGroup
	listener net.Listener
	shutdown chan struct{}
}

// NewServer 创建新服务器
func NewServer(addr string, initFunc func(ch *Channel)) (*Server, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	// 创建事件循环组
	group, err := NewEventLoopGroup(4, initFunc)
	if err != nil {
		return nil, err
	}

	return &Server{
		group:    group,
		listener: listener,
		shutdown: make(chan struct{}),
	}, nil
}

// Start 启动服务器
func (s *Server) Start() {
	// 为每个事件循环创建监听器
	for _, loop := range s.group.loops {
		go func(l *EventLoop) {
			if err := l.Run(s.listener); err != nil {
				slog.Error("EventLoop exited: ", "error", err)
			}
		}(loop)
	}

	slog.Info("Server started!", "port", s.listener.Addr())

	// 等待关闭信号
	select {
	case <-s.shutdown:
		slog.Info("Server shutdown")
	}
}

// Shutdown 关闭服务器
func (s *Server) Shutdown() {
	close(s.shutdown)
	s.group.Shutdown()
	s.listener.Close()
}
