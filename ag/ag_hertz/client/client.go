package client

import (
	"context"
	"errors"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
)

type Client struct {
	nc      naming_client.INamingClient
	hostUrl string
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

func NewClient(options []ClientOption) *Client {
	c := &Client{}
	for _, opt := range options {
		opt(c)
	}
	return c
}

func (c *Client) Invoke(ctx context.Context, method, path string, pathVars map[string]string, args any, reply any, opts ...config.RequestOption) error {
	var c1 *cli
	options := make([]Option, 0)
	if c.nc != nil {
		options = append(options, withNamingClient(c.nc))
		opts = append(opts, config.WithSD(true))
	}

	if c.hostUrl == "" {
		return errors.New("hostUrl is empty")
	}

	options = append(options, withHostUrl(c.hostUrl))
	c1, _ = newClient(getOptions(options...))
	_, err := c1.r().
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
