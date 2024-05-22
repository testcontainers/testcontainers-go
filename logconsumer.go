package testcontainers

import "github.com/testcontainers/testcontainers-go/log"

// Deprecated: use container.StdoutLog instead
// StdoutLog is the log type for STDOUT
const StdoutLog = "STDOUT"

// StderrLog is the log type for STDERR
const StderrLog = "STDERR"

// Deprecated: use log.Log instead
// Log represents a message that was created by a process,
// LogType is either "STDOUT" or "STDERR",
// Content is the byte contents of the message itself
type Log = log.Log

// logConsumerInterface {

// Deprecated: use container.LogConsumer instead
// LogConsumer represents any object that can
// handle a Log, it is up to the LogConsumer instance
// what to do with the log
type LogConsumer interface {
	Accept(Log)
}

// }

// LogConsumerConfig is a configuration object for the producer/consumer pattern
type LogConsumerConfig struct {
	Opts      []LogProductionOption // options for the production of logs
	Consumers []LogConsumer         // consumers for the logs
}
