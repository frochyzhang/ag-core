package nettypoll

const (
	miniNettyServerPropertiesPrefix = "mininetty.server"
	DefaultHertzOriginPort          = 8080
)

type MiniNettyServerProperties struct {
	Host          string `value:"${host:0.0.0.0}"`
	Port          int    `value:"${port:0}"`
	AdaptivePort  bool   `value:"${adaptive-port:false}"`
	ServiceName   string `value:"${service-name:}"`
	EnableIPRange string `value:"${enable-ip-range:}"`
}
