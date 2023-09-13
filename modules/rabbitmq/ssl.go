package rabbitmq

type SSLVerificationMode string

const (
	SSLVerificationModeNone SSLVerificationMode = "verify_none"
	SSLVerificationModePeer SSLVerificationMode = "verify_peer"
)

type SSLSettings struct {
	// Path to the CA certificate file
	CACertFile string
	// Path to the client certificate file
	CertFile string
	// Path to the key file
	KeyFile string
	// Verification mode
	VerificationMode SSLVerificationMode
	// Fail if no certificate is provided
	FailIfNoCert bool
	// Depth of certificate chain verification
	VerificationDepth int
}
