package rabbitmq_test

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"github.com/mdelapenya/tlscert"
)

func Example_tlsCertificates() {
	// exampleTLSCertificates {
	certDir, err := os.MkdirTemp("", "tls-example-")
	if err != nil {
		log.Print(err)
		return
	}
	defer os.RemoveAll(certDir)

	// Create a self-signed CA and use it to sign the server certificate.
	caCert := tlscert.SelfSignedFromRequest(tlscert.Request{
		Name: "ca", Host: "localhost", IsCA: true, ParentDir: certDir,
	})
	if caCert == nil {
		log.Print("failed to generate CA certificate")
		return
	}
	serverCert := tlscert.SelfSignedFromRequest(tlscert.Request{
		Name: "server", Host: "localhost,127.0.0.1", Parent: caCert, ParentDir: certDir,
	})
	if serverCert == nil {
		log.Print("failed to generate server certificate")
		return
	}

	keyPair, err := tls.LoadX509KeyPair(serverCert.CertPath, serverCert.KeyPath)
	if err != nil {
		log.Print(err)
		return
	}
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintln(w, "Hello over TLS!")
	}))
	server.TLS = &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{keyPair},
	}
	server.StartTLS()
	defer server.Close()

	// Trust the generated CA while retaining certificate and hostname verification.
	caPEM, err := os.ReadFile(caCert.CertPath)
	if err != nil {
		log.Print(err)
		return
	}
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(caPEM) {
		log.Print("failed to load CA certificate")
		return
	}
	transport := &http.Transport{TLSClientConfig: &tls.Config{
		MinVersion: tls.VersionTLS12,
		RootCAs:    roots,
	}}
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: transport, Timeout: 5 * time.Second}
	response, err := client.Get(server.URL)
	if err != nil {
		log.Print(err)
		return
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		log.Printf("unexpected response: %s", response.Status)
		return
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Print(err)
		return
	}
	fmt.Print(string(body))
	// }

	// Output: Hello over TLS!
}
