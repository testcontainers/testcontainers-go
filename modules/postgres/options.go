package postgres

type SSLVerificationMode string

type SSLSettings struct {
	// Path to the CA certificate file
	CACertFile string
	// Path to the client certificate file
	CertFile string
	// Path to the key file
	KeyFile string
	// Entrypoint used to override and set SSL ownership
	Entrypoint string
}
