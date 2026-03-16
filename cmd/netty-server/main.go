package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/frochyzhang/ag-core/ag/ag_netty"
)

func main() {
	addr := flag.String("addr", ":9090", "listen address")
	tlsMode := flag.String("tls", "tlcp", "TLS mode: none, tls, tlcp, auto")
	certFile := flag.String("cert", "", "TLS certificate file")
	keyFile := flag.String("key", "", "TLS private key file")
	signCert := flag.String("sign-cert", "certs/sign.crt", "TLCP signing certificate file")
	signKey := flag.String("sign-key", "certs/sign.key", "TLCP signing private key file")
	encCert := flag.String("enc-cert", "certs/enc.crt", "TLCP encryption certificate file")
	encKey := flag.String("enc-key", "certs/enc.key", "TLCP encryption private key file")
	caCert := flag.String("ca", "certs/ca.crt", "CA certificate file")
	insecure := flag.Bool("insecure", false, "skip TLS/TLCP certificate verification")
	flag.Parse()

	var opts []ag_netty.ServerOption

	if *tlsMode != "none" {
		mode, err := parseTLSMode(*tlsMode)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		opts = append(opts, ag_netty.WithTLS(&ag_netty.TLSConfig{
			Mode:               mode,
			CertFile:           *certFile,
			KeyFile:            *keyFile,
			SignCertFile:       *signCert,
			SignKeyFile:        *signKey,
			EncCertFile:        *encCert,
			EncKeyFile:         *encKey,
			CACertFile:         *caCert,
			InsecureSkipVerify: *insecure,
		}))
	}

	srv, err := ag_netty.NewServer(*addr, func(ch *ag_netty.Channel) {
		ch.Pipeline.AddLast("logging", ag_netty.NewLoggingHandler("server"))
		ch.Pipeline.AddLast("echo", &ag_netty.EchoHandler{})
	}, opts...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create server: %v\n", err)
		os.Exit(1)
	}

	srv.Start()
	slog.Info("netty echo server running", "addr", srv.Addr(), "tls", *tlsMode)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	slog.Info("shutting down...")
	srv.Shutdown()
	slog.Info("server stopped")
}

func parseTLSMode(s string) (ag_netty.TLSMode, error) {
	switch s {
	case "none":
		return ag_netty.TLSModeNone, nil
	case "tls":
		return ag_netty.TLSModeStandard, nil
	case "tlcp":
		return ag_netty.TLSModeTLCP, nil
	case "auto":
		return ag_netty.TLSModeAuto, nil
	default:
		return 0, fmt.Errorf("unknown TLS mode %q (use: none, tls, tlcp, auto)", s)
	}
}
