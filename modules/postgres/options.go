package postgres

type SSLVerificationMode string

const (
	SSLVerificationModeNone    SSLVerificationMode = "disable"
	SSLVerificationModeRequire SSLVerificationMode = "require"
)

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
