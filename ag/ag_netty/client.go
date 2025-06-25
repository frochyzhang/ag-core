package ag_netty

import (
	"sync"
	"time"
)

// Client 客户端
type Client struct {
	eventLoop      *ClientEventLoop
	channel        *Channel
	initFunc       func(ch *Channel)
	addr           string
	connectTimeout time.Duration
	readTimeout    time.Duration
	writeTimeout   time.Duration
	idleTimeout    time.Duration
	mu             sync.Mutex
}

// NewClient 创建新客户端
func NewClient(
	addr string,
	connectTimeout time.Duration,
	readTimeout time.Duration,
	writeTimeout time.Duration,
	idleTimeout time.Duration,
	initFunc func(ch *Channel),
) *Client {
	return &Client{
		addr:           addr,
		connectTimeout: connectTimeout,
		readTimeout:    readTimeout,
		writeTimeout:   writeTimeout,
		idleTimeout:    idleTimeout,
		initFunc:       initFunc,
	}
}

// Connect 连接到服务器
func (c *Client) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.eventLoop != nil {
		return nil // 已连接
	}

	// 创建事件循环
	eventLoop := NewClientEventLoop(c.initFunc)
	c.eventLoop = eventLoop

	// 建立连接
	channel, err := Dial(c.addr, c.connectTimeout, c.readTimeout, c.writeTimeout, c.idleTimeout, eventLoop)
	if err != nil {
		eventLoop.Shutdown()
		return err
	}

	c.channel = channel
	return nil
}

// Channel 获取通道
func (c *Client) Channel() *Channel {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.channel
}

// Close 关闭客户端
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.eventLoop != nil {
		c.eventLoop.Shutdown()
		//c.eventLoop = nil
		//c.channel = nil
		c.channel.Close()
	}
}

func (c *Client) Send(data []byte) {
	c.Channel().Write(data)
}
func (c *Client) SendAndGet(data []byte) (any, error) {
	future := c.Channel().WriteAsync(data)
	defer c.Close()
	ret, err := future.GetWithTimeout(c.readTimeout)
	return ret, err
}
