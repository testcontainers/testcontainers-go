package tls_test

import (
	stdlibtls "crypto/tls"
	"net"
	"os"
	"reflect"
	"testing"

	"github.com/testcontainers/testcontainers-go/tls"
)

func TestGenerate(t *testing.T) {
	t.Run("No host returns error", func(t *testing.T) {
		cert, err := tls.GenerateCert()
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !reflect.ValueOf(cert).IsZero() {
			t.Fatal("expected cert to be the zero value, got", cert)
		}
	})

	t.Run("With host", func(tt *testing.T) {
		cert, err := tls.GenerateCert(tls.WithHost("localhost"), tls.AsPem())
		if err != nil {
			tt.Fatal(err)
		}

		if reflect.ValueOf(cert).IsZero() {
			t.Fatal("expected cert not to be the zero value, got", cert)
		}

		if cert.Key == nil {
			tt.Fatal("expected key, got nil")
		}

		_, err = stdlibtls.X509KeyPair(cert.Bytes, cert.KeyBytes)
		if err != nil {
			tt.Fatal(err)
		}
	})

	t.Run("With multiple hosts", func(t *testing.T) {
		ip := "1.2.3.4"
		cert, err := tls.GenerateCert(tls.WithHost("localhost, " + ip))
		if err != nil {
			t.Fatal(err)
		}

		if cert.Key == nil {
			t.Fatal("expected key, got nil")
		}

		c := cert.Cert
		if len(c.IPAddresses) != 1 {
			t.Fatal("expected 1 IP address, got", len(c.IPAddresses))
		}

		if c.IPAddresses[0].String() != ip {
			t.Fatalf("expected IP address to be %s, got %s\n", ip, c.IPAddresses[0].String())
		}
	})

	t.Run("With multiple hosts and IPs", func(t *testing.T) {
		ip := "1.2.3.4"
		ips := []net.IP{net.ParseIP("4.5.6.7"), net.ParseIP("8.9.10.11")}
		cert, err := tls.GenerateCert(tls.WithHost("localhost, "+ip), tls.WithIPAddresses(ips...))
		if err != nil {
			t.Fatal(err)
		}

		if cert.Key == nil {
			t.Fatal("expected key, got nil")
		}

		c := cert.Cert
		if len(c.IPAddresses) != 3 {
			t.Fatal("expected 3 IP address, got", len(c.IPAddresses))
		}

		for i, ip := range ips {
			if c.IPAddresses[i].String() != ip.String() {
				t.Fatalf("expected IP address to be %s, got %s\n", ip.String(), c.IPAddresses[i].String())
			}
		}
		// the IP from the host comes after the IPs from the IPAddresses option
		if c.IPAddresses[2].String() != ip {
			t.Fatalf("expected IP address to be %s, got %s\n", ip, c.IPAddresses[0].String())
		}
	})

	t.Run("As CA", func(t *testing.T) {
		cert, err := tls.GenerateCert(tls.WithHost("localhost"), tls.AsCA())
		if err != nil {
			t.Fatal(err)
		}

		if cert.Cert == nil {
			t.Fatal("expected cert, got nil")
		}
		if cert.Key == nil {
			t.Fatal("expected key, got nil")
		}
		if cert.Bytes == nil {
			t.Fatal("expected bytes, got nil")
		}

		if !cert.Cert.IsCA {
			t.Fatal("expected cert to be CA, got false")
		}
	})

	t.Run("As PEM", func(t *testing.T) {
		cert, err := tls.GenerateCert(tls.WithHost("localhost"), tls.AsPem())
		if err != nil {
			t.Fatal(err)
		}

		// PEM format adds the key bytes to the cert struct
		if cert.Bytes == nil {
			t.Fatal("expected bytes, got nil")
		}
		if cert.KeyBytes == nil {
			t.Fatal("expected key bytes, got nil")
		}
	})

	t.Run("With Subject common name", func(t *testing.T) {
		cert, err := tls.GenerateCert(tls.WithHost("localhost"), tls.WithSubjectCommonName("Test"))
		if err != nil {
			t.Fatal(err)
		}

		if cert.Cert == nil {
			t.Fatal("expected cert, got nil")
		}

		c := cert.Cert
		if c.Subject.CommonName != "Test" {
			t.Fatal("expected common name to be Test, got", c.Subject.CommonName)
		}
	})

	t.Run("With Parent cert", func(t *testing.T) {
		parent, err := tls.GenerateCert(tls.WithHost("localhost"), tls.AsCA())
		if err != nil {
			t.Fatal(err)
		}

		cert, err := tls.GenerateCert(tls.WithHost("localhost"), tls.WithParent(parent))
		if err != nil {
			t.Fatal(err)
		}

		if cert.Cert == nil {
			t.Fatal("expected cert, got nil")
		}
		if cert.Key == nil {
			t.Fatal("expected key, got nil")
		}

		c := cert.Cert
		if c.Issuer.CommonName != parent.Cert.Subject.CommonName {
			t.Fatal("expected issuer to be parent, got", c.Issuer.CommonName)
		}
	})

	t.Run("With IP addresses", func(t *testing.T) {
		ip := "1.2.3.4"
		cert, err := tls.GenerateCert(tls.WithHost("localhost"), tls.WithIPAddresses(net.ParseIP(ip)))
		if err != nil {
			t.Fatal(err)
		}

		if cert.Cert == nil {
			t.Fatal("expected cert, got nil")
		}

		c := cert.Cert
		if len(c.IPAddresses) != 1 {
			t.Fatal("expected 1 IP address, got", len(c.IPAddresses))
		}

		if c.IPAddresses[0].String() != ip {
			t.Fatalf("expected IP address to be %s, got %s\n", ip, c.IPAddresses[0].String())
		}
	})

	t.Run("Save to file", func(tt *testing.T) {
		tmp := tt.TempDir()

		// no need to pass the AsPem option, the SaveToFile option will do that
		cert, err := tls.GenerateCert(tls.WithHost("localhost"), tls.WithSaveToFile(tmp))
		if err != nil {
			tt.Fatal(err)
		}

		inMemoryCert, err := stdlibtls.X509KeyPair(cert.Bytes, cert.KeyBytes)
		if err != nil {
			tt.Fatal(err)
		}

		// check if file existed
		certBytes, err := os.ReadFile(cert.CertPath)
		if err != nil {
			tt.Fatal(err)
		}

		certKeyBytes, err := os.ReadFile(cert.KeyPath)
		if err != nil {
			tt.Fatal(err)
		}

		fileCert, err := stdlibtls.X509KeyPair(certBytes, certKeyBytes)
		if err != nil {
			tt.Fatal(err)
		}

		for i, cert := range inMemoryCert.Certificate {
			if string(cert) != string(fileCert.Certificate[i]) {
				tt.Fatalf("expected certificate to be %s, got %s\n", string(cert), string(fileCert.Certificate[i]))
			}
		}
	})
}
