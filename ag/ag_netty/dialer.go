package ag_netty

import (
	"fmt"
	"log/slog"
	"net"
	"time"
)

// Dial 建立到服务器的连接（统一跨平台实现）
func Dial(
	addr string,
	connTimeout time.Duration,
	readTimeout time.Duration,
	writeTimeout time.Duration,
	idleTimeout time.Duration,
	looper EventLooper,
	initFunc func(ch *Channel),
	tlsConfig *TLSConfig,
) (*Channel, error) {
	var netConn net.Conn
	var err error

	if tlsConfig != nil && tlsConfig.Mode != TLSModeNone {
		netConn, err = tlsConfig.DialWithTLS(addr, connTimeout)
	} else {
		netConn, err = net.DialTimeout("tcp", addr, connTimeout)
	}
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", addr, err)
	}

	conn := NewNetConnAdapter(netConn)
	if writeTimeout > 0 {
		conn.SetWriteTimeout(writeTimeout)
	}

	channel := NewChannel(conn, looper)

	if initFunc != nil {
		initFunc(channel)
	}

	looper.Post(func() {
		channel.Pipeline.FireActive()
	})

	go readLoop(netConn, channel, looper, idleTimeout)

	slog.Info("connected to server", "addr", addr)
	return channel, nil
}

func readLoop(conn net.Conn, channel *Channel, looper EventLooper, idleTimeout time.Duration) {
	defer channel.Close()

	buffer := make([]byte, 4096)
	for {
		if looper.IsShutdown() {
			return
		}

		if idleTimeout > 0 {
			conn.SetReadDeadline(time.Now().Add(idleTimeout))
		}

		n, err := conn.Read(buffer)
		if err != nil {
			if !looper.IsShutdown() {
				looper.Post(func() {
					channel.Pipeline.FireError(err)
				})
			}
			return
		}

		if n > 0 {
			data := make([]byte, n)
			copy(data, buffer[:n])

			looper.Post(func() {
				channel.Pipeline.FireRead(data)
			})
		}
	}
}

type timeoutError struct {
	op      string
	addr    string
	timeout time.Duration
}

func (e *timeoutError) Error() string {
	return fmt.Sprintf("timeout: %s %s %s", e.op, e.addr, e.timeout)
}

func (e *timeoutError) Timeout() bool   { return true }
func (e *timeoutError) Temporary() bool { return true }
