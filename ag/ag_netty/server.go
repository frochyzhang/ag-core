package ag_netty

import (
	"log/slog"
	"net"
	"sync"
)

type Server struct {
	listener  net.Listener
	eventLoop *EventLoop
	initFunc  func(ch *Channel)
	connMap   sync.Map
	quit      chan struct{}
	wg        sync.WaitGroup
	tlsConfig *TLSConfig
}

type ServerOption func(*Server)

func WithTLS(cfg *TLSConfig) ServerOption {
	return func(s *Server) { s.tlsConfig = cfg }
}

func NewServer(addr string, initFunc func(ch *Channel), opts ...ServerOption) (*Server, error) {
	s := &Server{
		initFunc: initFunc,
		quit:     make(chan struct{}),
	}

	for _, opt := range opts {
		opt(s)
	}

	listener, err := s.createListener(addr)
	if err != nil {
		return nil, err
	}
	s.listener = listener
	s.eventLoop = NewEventLoop()

	return s, nil
}

func (s *Server) createListener(addr string) (net.Listener, error) {
	if s.tlsConfig != nil {
		return s.tlsConfig.CreateListener(addr)
	}
	return net.Listen("tcp", addr)
}

// Start 启动服务器（非阻塞，accept loop 在后台运行）
func (s *Server) Start() {
	s.wg.Add(1)
	go s.acceptLoop()
	slog.Info("server started", "addr", s.listener.Addr())
}

func (s *Server) acceptLoop() {
	defer s.wg.Done()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.quit:
				return
			default:
				slog.Error("accept error", "error", err)
				continue
			}
		}
		s.wg.Add(1)
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(netConn net.Conn) {
	defer s.wg.Done()

	conn := NewNetConnAdapter(netConn)
	channel := NewChannel(conn, s.eventLoop)
	s.connMap.Store(netConn, channel)

	if s.initFunc != nil {
		s.initFunc(channel)
	}

	s.eventLoop.Post(func() {
		channel.Pipeline.FireActive()
	})

	defer func() {
		s.connMap.Delete(netConn)
		channel.Close()
	}()

	buffer := make([]byte, 4096)
	for {
		if s.eventLoop.IsShutdown() {
			return
		}

		n, err := netConn.Read(buffer)
		if err != nil {
			if !s.eventLoop.IsShutdown() {
				s.eventLoop.Post(func() {
					channel.Pipeline.FireError(err)
				})
			}
			return
		}

		if n > 0 {
			data := make([]byte, n)
			copy(data, buffer[:n])

			s.eventLoop.Post(func() {
				channel.Pipeline.FireRead(data)
			})
		}
	}
}

func (s *Server) Shutdown() {
	close(s.quit)
	s.listener.Close()
	s.eventLoop.Shutdown()

	s.connMap.Range(func(key, value interface{}) bool {
		if ch, ok := value.(*Channel); ok {
			ch.Close()
		}
		return true
	})

	s.wg.Wait()
}

func (s *Server) Addr() net.Addr {
	return s.listener.Addr()
}
