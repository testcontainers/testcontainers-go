package vault

// LogLevel is a custom type that represents a logging level for the Vault
type LogLevel string

// The following constants define the possible logging levels for the Vault
// The default log level is info
const (
	Trace LogLevel = "trace"
	Debug LogLevel = "debug"
	Info  LogLevel = "info"
	Warn  LogLevel = "warn"
	Error LogLevel = "err"
)
