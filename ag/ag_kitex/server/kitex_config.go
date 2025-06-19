package server

const (
	KitexServerPropertiesPrefix = "kitex.server"
	DefaultKitexOriginPort      = 7000
)

type KitexServerProperties struct {
	Host string `value:"${:}"`
	Port int    `value:"${:0}"`
	// Port         int
	AdaptivePort  bool `value:"${:false}"`
	ServiceName   string
	EnableIPRange string `value:"${:}"`

	Grpc Grpc
}

type Grpc struct {
	Enable            bool `value:"${:false}"`
	MaxConnectionIdle int  `value:"${:0}"`
}
