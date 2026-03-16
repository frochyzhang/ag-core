package ag_netty

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"
	"time"

	"gitee.com/Trisia/gotlcp/pa"
	tlcppkg "gitee.com/Trisia/gotlcp/tlcp"
	"github.com/emmansun/gmsm/smx509"
)

type TLSMode int

const (
	TLSModeNone     TLSMode = iota // 明文 TCP
	TLSModeStandard                // 标准 TLS (1.2/1.3)
	TLSModeTLCP                    // 国密 TLCP (GB/T 38636-2020)
	TLSModeAuto                    // 自动识别：同端口同时支持 TLS 和 TLCP
)

type TLSConfig struct {
	Mode TLSMode

	// 标准 TLS 证书
	CertFile string
	KeyFile  string

	// TLCP 双证书（签名 + 加密）
	SignCertFile string
	SignKeyFile  string
	EncCertFile  string
	EncKeyFile   string

	// CA 根证书
	CACertFile string

	InsecureSkipVerify bool
}

// CreateListener 根据 TLS 配置创建监听器
func (c *TLSConfig) CreateListener(addr string) (net.Listener, error) {
	if c == nil || c.Mode == TLSModeNone {
		return net.Listen("tcp", addr)
	}

	switch c.Mode {
	case TLSModeStandard:
		cfg, err := c.buildStdTLSConfig()
		if err != nil {
			return nil, fmt.Errorf("build TLS config: %w", err)
		}
		return tls.Listen("tcp", addr, cfg)

	case TLSModeTLCP:
		cfg, err := c.buildTLCPConfig()
		if err != nil {
			return nil, fmt.Errorf("build TLCP config: %w", err)
		}
		return tlcppkg.Listen("tcp", addr, cfg)

	case TLSModeAuto:
		rawLn, err := net.Listen("tcp", addr)
		if err != nil {
			return nil, err
		}

		tlsCfg, err := c.buildStdTLSConfig()
		if err != nil {
			rawLn.Close()
			return nil, fmt.Errorf("build TLS config for auto: %w", err)
		}

		tlcpCfg, err := c.buildTLCPConfig()
		if err != nil {
			rawLn.Close()
			return nil, fmt.Errorf("build TLCP config for auto: %w", err)
		}

		return pa.NewListener(rawLn, tlcpCfg, tlsCfg), nil

	default:
		return nil, fmt.Errorf("unknown TLS mode: %d", c.Mode)
	}
}

// DialWithTLS 根据 TLS 配置建立连接
func (c *TLSConfig) DialWithTLS(addr string, timeout time.Duration) (net.Conn, error) {
	if c == nil || c.Mode == TLSModeNone {
		return net.DialTimeout("tcp", addr, timeout)
	}

	// 先建立 TCP 连接
	rawConn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return nil, err
	}

	switch c.Mode {
	case TLSModeStandard:
		cfg, err := c.buildStdTLSConfig()
		if err != nil {
			rawConn.Close()
			return nil, fmt.Errorf("build TLS config: %w", err)
		}
		tlsConn := tls.Client(rawConn, cfg)
		if err := tlsConn.Handshake(); err != nil {
			rawConn.Close()
			return nil, fmt.Errorf("TLS handshake: %w", err)
		}
		return tlsConn, nil

	case TLSModeTLCP:
		cfg, err := c.buildTLCPConfig()
		if err != nil {
			rawConn.Close()
			return nil, fmt.Errorf("build TLCP config: %w", err)
		}
		tlcpConn := tlcppkg.Client(rawConn, cfg)
		if err := tlcpConn.Handshake(); err != nil {
			rawConn.Close()
			return nil, fmt.Errorf("TLCP handshake: %w", err)
		}
		return tlcpConn, nil

	default:
		rawConn.Close()
		return nil, fmt.Errorf("unsupported TLS mode for client: %d (TLSModeAuto is server-only)", c.Mode)
	}
}

func (c *TLSConfig) buildStdTLSConfig() (*tls.Config, error) {
	cfg := &tls.Config{
		InsecureSkipVerify: c.InsecureSkipVerify,
		MinVersion:         tls.VersionTLS12,
	}

	if c.CertFile != "" && c.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(c.CertFile, c.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("load TLS keypair: %w", err)
		}
		cfg.Certificates = []tls.Certificate{cert}
	}

	if c.CACertFile != "" {
		pool, err := loadX509CertPool(c.CACertFile)
		if err != nil {
			return nil, err
		}
		cfg.RootCAs = pool
		cfg.ClientCAs = pool
	}

	return cfg, nil
}

func (c *TLSConfig) buildTLCPConfig() (*tlcppkg.Config, error) {
	cfg := &tlcppkg.Config{
		InsecureSkipVerify: c.InsecureSkipVerify,
	}

	if c.SignCertFile != "" && c.SignKeyFile != "" && c.EncCertFile != "" && c.EncKeyFile != "" {
		sigCertPEM, err := os.ReadFile(c.SignCertFile)
		if err != nil {
			return nil, fmt.Errorf("read sign cert: %w", err)
		}
		sigKeyPEM, err := os.ReadFile(c.SignKeyFile)
		if err != nil {
			return nil, fmt.Errorf("read sign key: %w", err)
		}
		encCertPEM, err := os.ReadFile(c.EncCertFile)
		if err != nil {
			return nil, fmt.Errorf("read enc cert: %w", err)
		}
		encKeyPEM, err := os.ReadFile(c.EncKeyFile)
		if err != nil {
			return nil, fmt.Errorf("read enc key: %w", err)
		}

		sigCert, err := tlcppkg.X509KeyPair(sigCertPEM, sigKeyPEM)
		if err != nil {
			return nil, fmt.Errorf("load TLCP sign keypair: %w", err)
		}
		encCert, err := tlcppkg.X509KeyPair(encCertPEM, encKeyPEM)
		if err != nil {
			return nil, fmt.Errorf("load TLCP enc keypair: %w", err)
		}

		cfg.Certificates = []tlcppkg.Certificate{sigCert, encCert}
	}

	if c.CACertFile != "" {
		pool, err := loadSMX509CertPool(c.CACertFile)
		if err != nil {
			return nil, err
		}
		cfg.RootCAs = pool
		cfg.ClientCAs = pool
	}

	return cfg, nil
}

func loadX509CertPool(caFile string) (*x509.CertPool, error) {
	caCert, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("read CA cert: %w", err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to parse CA certificate from %s", caFile)
	}
	return pool, nil
}

func loadSMX509CertPool(caFile string) (*smx509.CertPool, error) {
	caCert, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("read SM CA cert: %w", err)
	}
	pool := smx509.NewCertPool()
	if !pool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to parse SM CA certificate from %s", caFile)
	}
	return pool, nil
}
