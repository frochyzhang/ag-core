package server

import (
	"github.com/frochyzhang/ag-core/ag/ag_conf"
	"github.com/frochyzhang/ag-core/ag/ag_netty"
	"strings"
)

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

	// TLS/TLCP configuration
	TLSMode            string `value:"${tls.mode:none}"`            // none, tls, tlcp, auto
	TLSCertFile        string `value:"${tls.cert-file:}"`           // Standard TLS certificate
	TLSKeyFile         string `value:"${tls.key-file:}"`            // Standard TLS private key
	TLCPSignCertFile   string `value:"${tls.tlcp-sign-cert-file:}"` // TLCP signing certificate (SM2)
	TLCPSignKeyFile    string `value:"${tls.tlcp-sign-key-file:}"`  // TLCP signing private key (SM2)
	TLCPEncCertFile    string `value:"${tls.tlcp-enc-cert-file:}"`  // TLCP encryption certificate (SM2)
	TLCPEncKeyFile     string `value:"${tls.tlcp-enc-key-file:}"`   // TLCP encryption private key (SM2)
	CACertFile         string `value:"${tls.ca-cert-file:}"`        // CA certificate for peer verification
	InsecureSkipVerify bool   `value:"${tls.insecure-skip-verify:false}"`
}

func NewNettyServerProperties(binder ag_conf.IBinder) (*NettyServerProperties, error) {
	p := &NettyServerProperties{}
	err := binder.Bind(p, nettyServerPropertiesPrefix)
	return p, err
}

// TLSConfig converts properties to ag_netty.TLSConfig. Returns nil if TLS is disabled.
func (p *NettyServerProperties) TLSConfig() *ag_netty.TLSConfig {
	mode := parseTLSMode(p.TLSMode)
	if mode == ag_netty.TLSModeNone {
		return nil
	}
	return &ag_netty.TLSConfig{
		Mode:               mode,
		CertFile:           p.TLSCertFile,
		KeyFile:            p.TLSKeyFile,
		SignCertFile:       p.TLCPSignCertFile,
		SignKeyFile:        p.TLCPSignKeyFile,
		EncCertFile:        p.TLCPEncCertFile,
		EncKeyFile:         p.TLCPEncKeyFile,
		CACertFile:         p.CACertFile,
		InsecureSkipVerify: p.InsecureSkipVerify,
	}
}

func parseTLSMode(mode string) ag_netty.TLSMode {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "tls", "standard":
		return ag_netty.TLSModeStandard
	case "tlcp", "gmtls":
		return ag_netty.TLSModeTLCP
	case "auto", "adaptive":
		return ag_netty.TLSModeAuto
	default:
		return ag_netty.TLSModeNone
	}
}
