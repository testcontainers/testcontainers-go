package tls

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Certificate represents a certificate and private key pair. It's a wrapper
// around the x509.Certificate and rsa.PrivateKey types, and includes the raw
// bytes of the certificate and private key.
type Certificate struct {
	Cert     *x509.Certificate
	Bytes    []byte
	Key      *rsa.PrivateKey
	KeyBytes []byte
	CertPath string
	KeyPath  string
}

type certRequest struct {
	SubjectCommonName string            // CommonName is the subject name of the certificate
	Host              string            // Comma-separated hostnames and IPs to generate a certificate for
	ValidFrom         time.Time         // Creation date formatted as Jan 1 15:04:05 2011
	ValidFor          time.Duration     // Duration that certificate is valid for
	IsCA              bool              // whether this cert should be its own Certificate Authority
	IPAddreses        []net.IP          // IP addresses to include in the Subject Alternative Name
	Parent            *x509.Certificate // Parent is the parent certificate, if any
	Priv              any               // Priv is the private key of the parent certificate
	Pem               bool              // Whether to return the certificate as PEM bytes
	SaveTo            string            // Parent directory to save the certificate and private key
}

type CertOpt func(*certRequest)

func WithSubjectCommonName(commonName string) CertOpt {
	return func(r *certRequest) {
		r.SubjectCommonName = commonName
	}
}

// WithHost sets the hostnames and IPs to generate a certificate for.
// In the case the passed string contains comma-separated values,
// it will be split into multiple hostnames and IPs. Each hostname and IP
// will be trimmed of whitespace, and if the value is an IP, it will be
// added to the IPAddresses field of the certificate, after the ones
// passed with the WithIPAddresses option. Otherwise, it will be added
// to the DNSNames field.
func WithHost(host string) CertOpt {
	return func(r *certRequest) {
		r.Host = host
	}
}

func WithValidFrom(validFrom time.Time) CertOpt {
	return func(r *certRequest) {
		r.ValidFrom = validFrom
	}
}

func WithValidFor(validFor time.Duration) CertOpt {
	return func(r *certRequest) {
		r.ValidFor = validFor
	}
}

// AsCA sets the certificate as a Certificate Authority.
// When passed, the KeyUsage field of the certificate
// will append the x509.KeyUsageCertSign usage.
func AsCA() CertOpt {
	return func(r *certRequest) {
		r.IsCA = true
	}
}

// WithParent sets the parent certificate and private key of the certificate.
// It's used to sign the certificate with the parent certificate.
// At the moment the parent is set, the issuer of the certificate will be
// set to the common name of the parent certificate.
func WithParent(parent *x509.Certificate, priv any) CertOpt {
	return func(r *certRequest) {
		r.Parent = parent
		r.Priv = priv
	}
}

// AsPem sets the certificate to be returned as PEM bytes. It will include
// the private key in the KeyBytes field of the Certificate struct.
func AsPem() CertOpt {
	return func(r *certRequest) {
		r.Pem = true
	}
}

// WithIPAddresses sets the IP addresses of the certificate. They will be
// added first to the IPAddresses field of the certificate: those coming
// from the WithHost option will be added after these.
func WithIPAddresses(ips ...net.IP) CertOpt {
	return func(r *certRequest) {
		r.IPAddreses = append(r.IPAddreses, ips...)
	}
}

// WithSaveToFile sets the directory to save the certificate and private key.
// For that reason, it will set the AsPem option, as the certificate
// will be saved as PEM bytes, including the private key.
func WithSaveToFile(dir string) CertOpt {
	return func(r *certRequest) {
		r.SaveTo = dir

		if !r.Pem {
			AsPem()(r)
		}
	}
}

// newCertRequest returns a new certRequest with default values
// to avoid nil pointers.
func newCertRequest() certRequest {
	return certRequest{
		ValidFrom:  time.Now().Add(-time.Hour),
		ValidFor:   365 * 24 * time.Hour,
		IPAddreses: make([]net.IP, 0),
	}
}

// GenerateCert Generate a self-signed X.509 certificate for a TLS server. Returns
// a struct containing the certificate and private key, as well as the raw bytes
// of the certificate. In the case the  PEM option is set, the raw bytes will be
// PEM-encoded, including the bytes of the private key in the KeyBytes field.
func GenerateCert(opts ...CertOpt) (*Certificate, error) {
	req := newCertRequest()

	for _, opt := range opts {
		opt(&req)
	}

	if len(req.Host) == 0 {
		return nil, fmt.Errorf("missing required host")
	}

	keyUsage := x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature
	if req.IsCA {
		keyUsage |= x509.KeyUsageCertSign
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(2019),
		Subject: pkix.Name{
			CommonName: req.SubjectCommonName,
		},
		NotBefore:             req.ValidFrom,
		NotAfter:              req.ValidFrom.Add(req.ValidFor),
		KeyUsage:              keyUsage,
		BasicConstraintsValid: true,
		IsCA:                  req.IsCA,
	}

	if req.Parent == nil {
		// if no parent is provided, use the generated certificate as the parent
		req.Parent = &template
	} else {
		// if a parent is provided, use the parent's common name as the issuer
		template.Issuer.CommonName = req.Parent.Subject.CommonName
	}

	if len(req.IPAddreses) > 0 {
		template.IPAddresses = req.IPAddreses
	}

	hosts := strings.Split(req.Host, ",")
	for _, h := range hosts {
		h = strings.TrimSpace(h)
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	pk, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	if req.Priv == nil {
		// if no parent private key is provided, use the generated private key
		req.Priv = pk
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, &template, req.Parent, pk.Public(), req.Priv)
	if err != nil {
		return nil, err
	}

	cert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		return nil, err
	}

	certificate := &Certificate{
		Cert:  cert,
		Key:   pk,
		Bytes: certBytes,
	}

	if req.Pem {
		certificate.Bytes = pem.EncodeToMemory(&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: certBytes,
		})
		certificate.KeyBytes = pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(pk),
		})
	}

	if req.SaveTo != "" {
		id := uuid.NewString()
		certPath := filepath.Join(req.SaveTo, "cert-"+id+".pem")

		if err := os.WriteFile(certPath, certificate.Bytes, 0o644); err != nil {
			return nil, err
		}
		certificate.CertPath = certPath

		if certificate.KeyBytes != nil {
			keyPath := filepath.Join(req.SaveTo, "key-"+id+".pem")
			if err := os.WriteFile(keyPath, certificate.KeyBytes, 0o644); err != nil {
				return nil, err
			}
			certificate.KeyPath = keyPath
		}
	}

	return certificate, nil
}
