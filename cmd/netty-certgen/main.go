package main

import (
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/emmansun/gmsm/sm2"
	"github.com/emmansun/gmsm/smx509"
)

func main() {
	outDir := flag.String("out", "certs", "output directory for generated certificates")
	cn := flag.String("cn", "localhost", "common name and SAN DNS name")
	days := flag.Int("days", 3650, "certificate validity in days")
	flag.Parse()

	if err := os.MkdirAll(*outDir, 0755); err != nil {
		fatal("create output dir: %v", err)
	}

	caKey, caCertDER := generateCA(*days)
	caCert, err := smx509.ParseCertificate(caCertDER)
	if err != nil {
		fatal("parse CA cert: %v", err)
	}

	signKey := generateChildCert(*outDir, "sign", caCert, caKey, *cn, *days,
		x509.KeyUsageDigitalSignature,
		[]x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
	)
	_ = signKey

	encKey := generateChildCert(*outDir, "enc", caCert, caKey, *cn, *days,
		x509.KeyUsageKeyEncipherment|x509.KeyUsageDataEncipherment|x509.KeyUsageKeyAgreement,
		[]x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
	)
	_ = encKey

	writeKeyPEM(filepath.Join(*outDir, "ca.key"), caKey)
	writeCertPEM(filepath.Join(*outDir, "ca.crt"), caCertDER)

	fmt.Printf("generated %d files in %s/\n", 6, *outDir)
	fmt.Println("  ca.crt / ca.key       — SM2 CA (self-signed)")
	fmt.Println("  sign.crt / sign.key   — SM2 signing certificate")
	fmt.Println("  enc.crt / enc.key     — SM2 encryption certificate")
	fmt.Println()
	fmt.Println("TLCP server example:")
	fmt.Printf("  go run ./cmd/netty-server -tls tlcp -sign-cert %s/sign.crt -sign-key %s/sign.key -enc-cert %s/enc.crt -enc-key %s/enc.key -ca %s/ca.crt\n",
		*outDir, *outDir, *outDir, *outDir, *outDir)
	fmt.Println()
	fmt.Println("TLCP client example:")
	fmt.Printf("  go run ./cmd/netty-client -tls tlcp -sign-cert %s/sign.crt -sign-key %s/sign.key -enc-cert %s/enc.crt -enc-key %s/enc.key -ca %s/ca.crt -insecure\n",
		*outDir, *outDir, *outDir, *outDir, *outDir)
}

func generateCA(days int) (*sm2.PrivateKey, []byte) {
	key, err := sm2.GenerateKey(rand.Reader)
	if err != nil {
		fatal("generate CA key: %v", err)
	}

	tpl := smx509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "TLCP Test CA", Country: []string{"CN"}, Organization: []string{"ag-core"}},
		NotBefore:             time.Now().Add(-24 * time.Hour),
		NotAfter:              time.Now().AddDate(0, 0, days),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	der, err := smx509.CreateCertificate(rand.Reader, &tpl, &tpl, key.Public(), key)
	if err != nil {
		fatal("create CA cert: %v", err)
	}
	return key, der
}

func generateChildCert(
	outDir, name string,
	caCert *smx509.Certificate, caKey *sm2.PrivateKey,
	cn string, days int,
	keyUsage x509.KeyUsage,
	extKeyUsage []x509.ExtKeyUsage,
) *sm2.PrivateKey {
	key, err := sm2.GenerateKey(rand.Reader)
	if err != nil {
		fatal("generate %s key: %v", name, err)
	}

	tpl := smx509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject:      pkix.Name{CommonName: cn + "_" + name, Country: []string{"CN"}, Organization: []string{"ag-core"}},
		NotBefore:    time.Now().Add(-24 * time.Hour),
		NotAfter:     time.Now().AddDate(0, 0, days),
		KeyUsage:     keyUsage,
		ExtKeyUsage:  extKeyUsage,
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
		DNSNames:     []string{"localhost", cn},
	}

	der, err := smx509.CreateCertificate(rand.Reader, &tpl, caCert, key.Public(), caKey)
	if err != nil {
		fatal("create %s cert: %v", name, err)
	}

	writeKeyPEM(filepath.Join(outDir, name+".key"), key)
	writeCertPEM(filepath.Join(outDir, name+".crt"), der)
	return key
}

func writeKeyPEM(path string, key *sm2.PrivateKey) {
	der, err := smx509.MarshalSM2PrivateKey(key)
	if err != nil {
		fatal("marshal key: %v", err)
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		fatal("create %s: %v", path, err)
	}
	defer f.Close()
	if err := pem.Encode(f, &pem.Block{Type: "SM2 PRIVATE KEY", Bytes: der}); err != nil {
		fatal("encode key PEM: %v", err)
	}
}

func writeCertPEM(path string, der []byte) {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fatal("create %s: %v", path, err)
	}
	defer f.Close()
	if err := pem.Encode(f, &pem.Block{Type: "CERTIFICATE", Bytes: der}); err != nil {
		fatal("encode cert PEM: %v", err)
	}
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
