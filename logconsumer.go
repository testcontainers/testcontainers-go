package testcontainers

import (
	"github.com/testcontainers/testcontainers-go/log"
)

// Deprecated: use log.Stdout instead
// StdoutLog is the log type for STDOUT
const StdoutLog = log.Stdout

// Deprecated: use log.Stderr instead
// StderrLog is the log type for STDERR
const StderrLog = log.Stderr

// Deprecated: use log.Log instead
// Log represents a message that was created by a process,
// LogType is either "STDOUT" or "STDERR",
// Content is the byte contents of the message itself
type Log = log.Log

// logConsumerInterface {

// Deprecated: use log.Consumer instead
// LogConsumer represents any object that can
// handle a Log, it is up to the LogConsumer instance
// what to do with the log
type LogConsumer = log.Consumer

// }

// Deprecated: use container.LogConsumerConfig instead
// LogConsumerConfig is a configuration object for the producer/consumer pattern
type LogConsumerConfig = log.ConsumerConfig
