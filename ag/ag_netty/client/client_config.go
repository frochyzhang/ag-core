package client

import (
	"github.com/frochyzhang/ag-core/ag/ag_netty"
	"strings"
)

const (
	NettyClientPropertiesPrefix = "netty.client"
)

type NettyClientProperties struct {
	Addr           string `value:"${addr}"`
	ConnectTimeout int    `value:"${connect-timeout:50}"`
	ReadTimeout    int    `value:"${read-timeout:200}"`
	WriteTimeout   int    `value:"${write-timeout:200}"`
	IdleTimeout    int    `value:"${idle-timeout:10000}"`

	// TLS/TLCP configuration
	TLSMode            string `value:"${tls.mode:none}"`            // none, tls, tlcp
	TLSCertFile        string `value:"${tls.cert-file:}"`           // Client certificate (mutual TLS)
	TLSKeyFile         string `value:"${tls.key-file:}"`            // Client private key (mutual TLS)
	TLCPSignCertFile   string `value:"${tls.tlcp-sign-cert-file:}"` // TLCP signing certificate (SM2)
	TLCPSignKeyFile    string `value:"${tls.tlcp-sign-key-file:}"`  // TLCP signing private key (SM2)
	TLCPEncCertFile    string `value:"${tls.tlcp-enc-cert-file:}"`  // TLCP encryption certificate (SM2)
	TLCPEncKeyFile     string `value:"${tls.tlcp-enc-key-file:}"`   // TLCP encryption private key (SM2)
	CACertFile         string `value:"${tls.ca-cert-file:}"`        // CA certificate for server verification
	InsecureSkipVerify bool   `value:"${tls.insecure-skip-verify:false}"`
}

// TLSConfig converts properties to ag_netty.TLSConfig. Returns nil if TLS is disabled.
func (p *NettyClientProperties) TLSConfig() *ag_netty.TLSConfig {
	mode := parseClientTLSMode(p.TLSMode)
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

func parseClientTLSMode(mode string) ag_netty.TLSMode {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "tls", "standard":
		return ag_netty.TLSModeStandard
	case "tlcp", "gmtls":
		return ag_netty.TLSModeTLCP
	default:
		return ag_netty.TLSModeNone
	}
}
