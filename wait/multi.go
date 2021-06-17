package wait

import (
	"context"
	"fmt"
	"time"
)

// Implement interface
var _ Strategy = (*MultiStrategy)(nil)

type MultiStrategy struct {
	// all Strategies should have a timeout to avoid waiting infinitely
	timeout time.Duration

	// additional properties
	Strategies []Strategy
}

// WithStartupTimeout can be used to change the default startup timeout
//
// Deprecated: use WithTimeout instead
func (s *MultiStrategy) WithStartupTimeout(timeout time.Duration) *MultiStrategy {
	return s.WithTimeout(timeout)
}

// WithTimeout can be used to change the default startup timeout
func (ms *MultiStrategy) WithTimeout(timeout time.Duration) *MultiStrategy {
	ms.timeout = timeout
	return ms
}

func ForAll(strategies ...Strategy) *MultiStrategy {
	return &MultiStrategy{
		timeout:    defaultTimeout(),
		Strategies: strategies,
	}
}

func (ms *MultiStrategy) WaitUntilReady(ctx context.Context, target StrategyTarget) (err error) {
	ctx, cancelContext := context.WithTimeout(ctx, ms.timeout)
	defer cancelContext()

	if len(ms.Strategies) == 0 {
		return fmt.Errorf("no wait strategy supplied")
	}

	for _, strategy := range ms.Strategies {
		err := strategy.WaitUntilReady(ctx, target)
		if err != nil {
			return err
		}
	}
	return nil
}
