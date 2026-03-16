package ag_netty

import (
	"net"
	"time"
)

// Connection 网络连接接口
type Connection interface {
	// Read reads data from the connection
	Read() ([]byte, error)

	// Write writes data to the connection
	Write(data []byte) (int, error)

	// Close closes the connection
	Close() error

	// RemoteAddr returns the remote network address
	RemoteAddr() net.Addr

	// LocalAddr returns the local network address
	LocalAddr() net.Addr

	// SetReadTimeout sets the read timeout
	SetReadTimeout(timeout time.Duration)

	// SetWriteTimeout sets the write timeout
	SetWriteTimeout(timeout time.Duration)

	// SetIdleTimeout sets the idle timeout
	SetIdleTimeout(timeout time.Duration)

	// NetConn returns the underlying net.Conn
	NetConn() net.Conn
}

// NetConnAdapter net.Conn 适配器
type NetConnAdapter struct {
	conn         net.Conn
	writeTimeout time.Duration
}

// NewNetConnAdapter 创建 net.Conn 适配器
func NewNetConnAdapter(conn net.Conn) Connection {
	return &NetConnAdapter{conn: conn}
}

func (a *NetConnAdapter) Read() ([]byte, error) {
	buf := make([]byte, 4096)
	n, err := a.conn.Read(buf)
	if err != nil {
		return nil, err
	}
	data := make([]byte, n)
	copy(data, buf[:n])
	return data, nil
}

func (a *NetConnAdapter) Write(data []byte) (int, error) {
	if a.writeTimeout > 0 {
		a.conn.SetWriteDeadline(time.Now().Add(a.writeTimeout))
	}
	return a.conn.Write(data)
}

func (a *NetConnAdapter) Close() error {
	return a.conn.Close()
}

func (a *NetConnAdapter) RemoteAddr() net.Addr {
	return a.conn.RemoteAddr()
}

func (a *NetConnAdapter) LocalAddr() net.Addr {
	return a.conn.LocalAddr()
}

func (a *NetConnAdapter) SetReadTimeout(timeout time.Duration) {
	if timeout > 0 {
		a.conn.SetReadDeadline(time.Now().Add(timeout))
	}
}

func (a *NetConnAdapter) SetWriteTimeout(timeout time.Duration) {
	a.writeTimeout = timeout
}

func (a *NetConnAdapter) SetIdleTimeout(timeout time.Duration) {
	if timeout > 0 {
		a.conn.SetDeadline(time.Now().Add(timeout))
	}
}

func (a *NetConnAdapter) NetConn() net.Conn {
	return a.conn
}
