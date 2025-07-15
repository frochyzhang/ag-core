package client

import (
	"context"
	"errors"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"time"

	hertzclient "github.com/cloudwego/hertz/pkg/app/client"
)

type Client struct {
	nc      naming_client.INamingClient
	hostUrl string
	*cli
	reqOpt []config.RequestOption
}
type ClientOption func(*Client)

func WithNamingClient(nc naming_client.INamingClient) ClientOption {
	return func(c *Client) {
		c.nc = nc
	}
}

func WithHostUrl(hostUrl string) ClientOption {
	return func(c *Client) {
		c.hostUrl = hostUrl
	}
}

func NewClient(opts []ClientOption) *Client {
	c := &Client{}
	for _, opt := range opts {
		opt(c)
	}

	options := make([]Option, 0)
	if c.nc != nil {
		options = append(options, withNamingClient(c.nc))
		c.reqOpt = append(c.reqOpt, config.WithSD(true))
	}

	if c.hostUrl == "" {
		panic(errors.New("hostUrl is empty"))
	}

	options = append(options, withHostUrl(c.hostUrl))

	// Add connection pooling options
	options = append(options, WithHertzClientOption(
		hertzclient.WithDialTimeout(5*time.Second),
		hertzclient.WithKeepAlive(true),
		hertzclient.WithMaxConnDuration(time.Minute),
		hertzclient.WithMaxConnWaitTimeout(5*time.Second),
		hertzclient.WithMaxConnsPerHost(100),
		hertzclient.WithMaxIdleConnDuration(90*time.Second),
	))

	client, err := newClient(getOptions(options...))
	if err != nil {
		panic(err)
	}
	c.cli = client

	return c
}

func (c *Client) Invoke(ctx context.Context, method, path string, pathVars map[string]string, args any, reply any, opts ...config.RequestOption) error {
	opts = append(c.reqOpt, opts...)
	_, err := c.cli.r().
		setContext(ctx).
		setQueryParams(map[string]interface{}{}).
		setPathParams(pathVars).
		addHeaders(map[string]string{}).
		setFormParams(map[string]string{}).
		setFormFileParams(map[string]string{}).
		setBodyParam(args).
		setRequestOption(opts...).
		setResult(reply).
		execute(method, path)
	if err != nil {
		return err
	}
	return nil
}
