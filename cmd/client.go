package main

import (
	"ag-core/ag/ag_netty"
	"ag-core/ag/ag_netty/client"
	"log/slog"
)

func main() {
	opts := make([]client.Option, 0)
	opts = append(opts, client.WithAddr("127.0.0.1:9090"))
	opts = append(opts, client.AppendHandler(&ag_netty.ConnectorHandler{}))
	opts = append(opts, client.AppendHandler(ag_netty.NewLoggingHandler("client")))
	opts = append(opts, client.AppendHandler(&client.EchoHandler{EchoHandler: &ag_netty.EchoHandler{}}))

	suite := &client.NettyOptionSuite{
		Opts: opts,
	}
	clientWithOpts := client.NewNettyClientWithSuite(suite, &slog.Logger{})

	err := clientWithOpts.Connect()
	if err != nil {
		slog.Error("Connection failed", "error", err)
		return
	}

	clientWithOpts.Send([]byte("hello ag server"))
	select {}
}
