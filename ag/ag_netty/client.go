package ag_netty

import (
	"sync"
	"time"
)

type Client struct {
	eventLoop      *EventLoop
	channel        *Channel
	initFunc       func(ch *Channel)
	addr           string
	connectTimeout time.Duration
	readTimeout    time.Duration
	writeTimeout   time.Duration
	idleTimeout    time.Duration
	tlsConfig      *TLSConfig
	mu             sync.Mutex
}

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

func (c *Client) SetTLSConfig(cfg *TLSConfig) {
	c.tlsConfig = cfg
}

func (c *Client) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.eventLoop != nil {
		return nil
	}

	eventLoop := NewEventLoop()
	c.eventLoop = eventLoop

	channel, err := Dial(
		c.addr,
		c.connectTimeout,
		c.readTimeout,
		c.writeTimeout,
		c.idleTimeout,
		eventLoop,
		c.initFunc,
		c.tlsConfig,
	)
	if err != nil {
		eventLoop.Shutdown()
		c.eventLoop = nil
		return err
	}

	c.channel = channel
	return nil
}

func (c *Client) Channel() *Channel {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.channel
}

func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.channel != nil {
		c.channel.Close()
		c.channel = nil
	}

	if c.eventLoop != nil {
		c.eventLoop.Shutdown()
		c.eventLoop = nil
	}
}

func (c *Client) Send(data []byte) {
	c.Channel().Write(data)
}

// SendAndGet 发送数据并等待响应。调用方负责管理 Client 生命周期（调用 Close）。
func (c *Client) SendAndGet(data []byte) (any, error) {
	future := c.Channel().WriteAsync(data)
	return future.GetWithTimeout(c.readTimeout)
}
