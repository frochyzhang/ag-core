package ag_netty

import (
	"io"
	"net"
	"sync"
)

// Channel 网络通道
type Channel struct {
	conn      Connection
	looper    EventLooper
	Pipeline  *Pipeline
	mu        sync.RWMutex // protects active flag and future
	active    bool
	closeOnce sync.Once
	future    *Future
}

// NewChannel 创建新通道
func NewChannel(conn Connection, looper EventLooper) *Channel {
	ch := &Channel{
		conn:   conn,
		looper: looper,
		active: true,
	}
	ch.Pipeline = NewPipeline(ch)
	return ch
}

func (c *Channel) Future() *Future {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.future
}

// Write 写数据
func (c *Channel) Write(data []byte) {
	c.looper.Post(func() {
		c.mu.RLock()
		active := c.active
		c.mu.RUnlock()
		if active {
			c.Pipeline.FireWrite(data)
		}
	})
}

// WriteDirect 直接写数据（无流水线处理）
func (c *Channel) WriteDirect(data []byte) error {
	c.mu.RLock()
	active := c.active
	c.mu.RUnlock()
	if !active {
		return io.ErrClosedPipe
	}
	_, err := c.conn.Write(data)
	return err
}

// WriteAsync 异步写数据
func (c *Channel) WriteAsync(data []byte) *Future {
	future := NewFuture()
	c.mu.Lock()
	c.future = future
	c.mu.Unlock()

	c.looper.Post(func() {
		c.mu.RLock()
		active := c.active
		c.mu.RUnlock()
		if active {
			c.Pipeline.FireWrite(data)
			// future is completed by response handler via ctx.Channel().Future().Complete()
		} else {
			future.Complete(io.ErrClosedPipe)
		}
	})
	return future
}

// Close 关闭通道
func (c *Channel) Close() {
	c.closeOnce.Do(func() {
		c.mu.Lock()
		c.active = false
		c.mu.Unlock()

		c.conn.Close()
		c.Pipeline.FireInactive()
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
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.active
}
