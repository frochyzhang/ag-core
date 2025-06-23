package ag_netpoll

import (
	"github.com/cloudwego/netpoll"
	"io"
	"net"
	"sync"
)

// Channel 网络通道
type Channel struct {
	conn      netpoll.Connection
	loop      *EventLoop
	Pipeline  *Pipeline
	active    bool
	closeOnce sync.Once
}

// NewChannel 创建新通道
func NewChannel(conn netpoll.Connection, loop *EventLoop) *Channel {
	ch := &Channel{
		conn:   conn,
		loop:   loop,
		active: true,
	}
	ch.Pipeline = NewPipeline(ch)
	return ch
}

// Write 写数据
func (c *Channel) Write(data []byte) {
	c.loop.Post(func() {
		if c.active {
			c.Pipeline.FireWrite(data)
		}
	})
}

// WriteDirect 直接写数据（无流水线处理）
func (c *Channel) WriteDirect(data []byte) error {
	if !c.active {
		return io.ErrClosedPipe
	}
	_, err := c.conn.Write(data)
	return err
}

// WriteAsync 异步写数据
func (c *Channel) WriteAsync(data []byte) *Future {
	future := NewFuture()
	c.loop.Post(func() {
		if c.active {
			_, err := c.conn.Write(data)
			future.Complete(err)
		} else {
			future.Complete(io.ErrClosedPipe)
		}
	})
	return future
}

// Close 关闭通道
func (c *Channel) Close() {
	c.closeOnce.Do(func() {
		c.loop.Post(func() {
			if c.active {
				c.active = false
				c.conn.Close()
				c.Pipeline.FireInactive()
			}
		})
	})
}

// RemoteAddr 获取远程地址
func (c *Channel) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

// LocalAddr 获取本地地址
func (c *Channel) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

// IsActive 检查通道是否活跃
func (c *Channel) IsActive() bool {
	return c.active
}
