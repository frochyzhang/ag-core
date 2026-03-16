package main

import (
	"bufio"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/frochyzhang/ag-core/ag/ag_netty"
)

type echoHandler struct{}

func (h *echoHandler) HandleActive(ctx *ag_netty.HandlerContext) {
	slog.Info("connected", "remote", ctx.Channel().RemoteAddr())
}
func (h *echoHandler) HandleInactive(ctx *ag_netty.HandlerContext) {
	slog.Info("disconnected")
}
func (h *echoHandler) HandleRead(ctx *ag_netty.HandlerContext, data []byte) {
	ctx.Channel().Future().Complete(string(data))
}
func (h *echoHandler) HandleWrite(ctx *ag_netty.HandlerContext, data []byte) {
	if err := ctx.Channel().WriteDirect(data); err != nil {
		slog.Error("WriteDirect failed", "error", err, "dataLen", len(data))
	}
}
func (h *echoHandler) HandleError(ctx *ag_netty.HandlerContext, err error) {
	slog.Error("connection error", "error", err)
}

func main() {
	addr := flag.String("addr", "127.0.0.1:9090", "server address")
	timeout := flag.Duration("timeout", 5*time.Second, "connect/read/write timeout")
	tlsMode := flag.String("tls", "tlcp", "TLS mode: none, tls, tlcp")
	certFile := flag.String("cert", "", "TLS client certificate file")
	keyFile := flag.String("key", "", "TLS client private key file")
	signCert := flag.String("sign-cert", "certs/sign.crt", "TLCP signing certificate file")
	signKey := flag.String("sign-key", "certs/sign.key", "TLCP signing private key file")
	encCert := flag.String("enc-cert", "certs/enc.crt", "TLCP encryption certificate file")
	encKey := flag.String("enc-key", "certs/enc.key", "TLCP encryption private key file")
	caCert := flag.String("ca", "certs/ca.crt", "CA certificate file")
	insecure := flag.Bool("insecure", false, "skip TLS/TLCP certificate verification")
	flag.Parse()

	client := ag_netty.NewClient(*addr, *timeout, *timeout, *timeout, 0, func(ch *ag_netty.Channel) {
		ch.Pipeline.AddLast("logging", ag_netty.NewLoggingHandler("client"))
		ch.Pipeline.AddLast("echo", &echoHandler{})
	})
	defer client.Close()

	if *tlsMode != "none" {
		mode, err := parseTLSMode(*tlsMode)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		client.SetTLSConfig(&ag_netty.TLSConfig{
			Mode:               mode,
			CertFile:           *certFile,
			KeyFile:            *keyFile,
			SignCertFile:       *signCert,
			SignKeyFile:        *signKey,
			EncCertFile:        *encCert,
			EncKeyFile:         *encKey,
			CACertFile:         *caCert,
			InsecureSkipVerify: *insecure,
		})
	}

	if err := client.Connect(); err != nil {
		fmt.Fprintf(os.Stderr, "connect failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("connected to", *addr)
	fmt.Println("type a message and press Enter (Ctrl+C to quit):")

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		result, err := client.SendAndGet([]byte(line))
		if err != nil {
			fmt.Fprintf(os.Stderr, "send error: %v\n", err)
			continue
		}
		fmt.Printf("< %v\n", result)
	}
}

func parseTLSMode(s string) (ag_netty.TLSMode, error) {
	switch s {
	case "none":
		return ag_netty.TLSModeNone, nil
	case "tls":
		return ag_netty.TLSModeStandard, nil
	case "tlcp":
		return ag_netty.TLSModeTLCP, nil
	default:
		return 0, fmt.Errorf("unknown TLS mode %q (use: none, tls, tlcp)", s)
	}
}
