package client

import (
	"ag-core/ag/ag_netty"
	"fmt"
	"log/slog"
)

type Client struct {
	*ag_netty.Client
	addr     string
	handlers []ag_netty.ChannelHandler
	logger   *slog.Logger
}

type Option struct {
	opt func(client *Client)
}

func WithAddr(addr string) Option {
	return Option{
		opt: func(c *Client) {
			c.addr = addr
		},
	}
}
func AppendHandler(ch ag_netty.ChannelHandler) Option {
	return Option{
		opt: func(c *Client) {
			c.handlers = append(c.handlers, ch)
		},
	}
}

func newClient(logger *slog.Logger, opts ...Option) *Client {
	c := &Client{
		handlers: make([]ag_netty.ChannelHandler, 0),
		logger:   logger,
	}

	for _, opt := range opts {
		opt.opt(c)
	}

	initFunc := func(ch *ag_netty.Channel) {
		pipeline := ch.Pipeline
		if pipeline != nil {
			for i, handler := range c.handlers {
				pipeline.AddLast(fmt.Sprintf("handler%d", i), handler)
			}
		}
	}

	client := ag_netty.NewClient(c.addr, initFunc)
	c.Client = client
	return c
}

type NettyOptionSuite struct {
	Opts []Option
}

func (s *NettyOptionSuite) options() []Option { return s.Opts }

func NewNettyClientWithSuite(
	suite *NettyOptionSuite,
	logger *slog.Logger,
) *Client {
	return newClient(logger, suite.options()...)
}

// EchoHandler 回显处理器(复用服务器端实现)
type EchoHandler struct {
	*ag_netty.EchoHandler
}

func (h *EchoHandler) HandleRead(ctx *ag_netty.HandlerContext, data []byte) {
	defer ctx.Close()
	slog.Info("Received response", "data", string(data))
}
