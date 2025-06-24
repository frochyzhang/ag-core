package server

const (
	nettyServerPropertiesPrefix = "netty.server"
	DefaultNettyOriginPort      = 8080
)

type NettyServerProperties struct {
	Host          string `value:"${host:0.0.0.0}"`
	Port          int    `value:"${port:0}"`
	AdaptivePort  bool   `value:"${adaptive-port:false}"`
	ServiceName   string `value:"${service-name:}"`
	EnableIPRange string `value:"${enable-ip-range:}"`
}
