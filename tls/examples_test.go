package tls_test

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	tctls "github.com/testcontainers/testcontainers-go/tls"
)

func ExampleGenerateCert() {
	tmp := os.TempDir()
	certsDir := tmp + "/certs"
	defer os.RemoveAll(certsDir)

	if err := os.MkdirAll(certsDir, 0755); err != nil {
		log.Fatal(err)
	}

	// Generate a certificate for localhost and save it to disk.
	// There is no need to pass the AsPem option: the SaveToFile option will automatically save the certificate as PEM.
	caCert, err := tctls.GenerateCert(tctls.WithHost("localhost"), tctls.WithSaveToFile(certsDir))
	if err != nil {
		log.Fatal(err)
	}

	cert, err := tctls.GenerateCert(tctls.WithHost("localhost"), tctls.WithSaveToFile(certsDir), tctls.WithParent(caCert.Cert, caCert.Key))
	if err != nil {
		log.Fatal(err)
	}

	// create an http server that uses the generated certificate
	// and private key to serve requests over HTTPS

	server := &http.Server{
		Addr: ":8443",
	}

	server.Handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("TLS works!\n"))
	})

	go func() {
		_ = server.ListenAndServeTLS(cert.CertPath, cert.KeyPath)
	}()
	defer server.Close()

	// perform an HTTP request to the server, using the generated certificate

	caCertPool := x509.NewCertPool()
	caCertPool.AddCert(cert.Cert)

	tlsConfig := &tls.Config{
		RootCAs: caCertPool,
	}

	tr := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	const url = "https://localhost:8443/hello"

	client := &http.Client{Transport: tr}
	resp, err := client.Get(url)
	if err != nil {
		log.Fatalf("Failed to get response: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}

	fmt.Println(string(body))

	// Output:
	// TLS works!

}
