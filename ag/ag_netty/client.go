package ag_netty

import (
	"sync"
)

// Client 客户端
type Client struct {
	eventLoop *ClientEventLoop
	channel   *Channel
	initFunc  func(ch *Channel)
	addr      string
	mu        sync.Mutex
}

// NewClient 创建新客户端
func NewClient(addr string, initFunc func(ch *Channel)) *Client {
	return &Client{
		addr:     addr,
		initFunc: initFunc,
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
	channel, err := Dial(c.addr, eventLoop)
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
		c.eventLoop = nil
		c.channel = nil
	}
}

func (c *Client) Send(data []byte) {
	c.Channel().Write(data)
}
