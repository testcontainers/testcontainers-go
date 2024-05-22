package log

import "time"

const StoppedForOutOfSyncMessage = "Stopping log consumer: Headers out of sync"

// logConsumerInterface {

// Consumer represents any object that can
// handle a Log, it is up to the Consumer instance
// what to do with the log
type Consumer interface {
	Accept(Log)
}

// }

// ConsumerConfig is a configuration object for the producer/consumer pattern
type ConsumerConfig struct {
	Opts      []ProductionOption // options for the production of logs
	Consumers []Consumer         // consumers for the logs
}

type OptionsContainer interface {
	WithLogProductionTimeout(timeout time.Duration)
}

// ProductionOption is a functional option that can be used to configure the log production
type ProductionOption func(OptionsContainer)

// WithProductionTimeout is a functional option that sets the timeout for the log production.
// If the timeout is lower than 5s or greater than 60s it will be set to 5s or 60s respectively.
func WithProductionTimeout(timeout time.Duration) ProductionOption {
	return func(c OptionsContainer) {
		c.WithLogProductionTimeout(timeout)
	}
}
