package client

const (
	NettyClientPropertiesPrefix = "netty.client"
)

type NettyClientProperties struct {
	Addr           string `value:"${addr}"`
	ConnectTimeout int    `value:"${connect-timeout:50}"`
	ReadTimeout    int    `value:"${read-timeout:200}"`
	WriteTimeout   int    `value:"${write-timeout:200}"`
	IdleTimeout    int    `value:"${idle-timeout:10000}"`
}
