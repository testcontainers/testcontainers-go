package wait

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMultiStrategy_WaitUntilReady(t *testing.T) {
	t.Parallel()
	type args struct {
		ctx    context.Context
		target StrategyTarget
	}
	tests := []struct {
		name     string
		strategy Strategy
		args     args
		wantErr  bool
	}{
		{
			name:     "returns error when no WaitStrategies are passed",
			strategy: ForAll(),
			args: args{
				ctx:    context.Background(),
				target: NopStrategyTarget{},
			},
			wantErr: true,
		},
		{
			name: "returns WaitStrategy error",
			strategy: ForAll(
				ForNop(
					func(_ context.Context, _ StrategyTarget) error {
						return errors.New("intentional failure")
					},
				),
			),
			args: args{
				ctx:    context.Background(),
				target: NopStrategyTarget{},
			},
			wantErr: true,
		},
		{
			name: "WithDeadline sets context Deadline for WaitStrategy",
			strategy: ForAll(
				ForNop(
					func(ctx context.Context, _ StrategyTarget) error {
						if _, set := ctx.Deadline(); !set {
							return errors.New("expected context.Deadline to be set")
						}
						return nil
					},
				),
				ForLog("docker"),
			).WithDeadline(1 * time.Second),
			args: args{
				ctx: context.Background(),
				target: NopStrategyTarget{
					ReaderCloser: io.NopCloser(bytes.NewReader([]byte("docker"))),
				},
			},
			wantErr: false,
		},
		{
			name: "WithStartupTimeoutDefault skips setting context.Deadline when WaitStrategy.Timeout is defined",
			strategy: ForAll(
				ForNop(
					func(ctx context.Context, _ StrategyTarget) error {
						if _, set := ctx.Deadline(); set {
							return errors.New("expected context.Deadline not to be set")
						}
						return nil
					},
				).WithStartupTimeout(2*time.Second),
				ForLog("docker"),
			).WithStartupTimeoutDefault(1 * time.Second),
			args: args{
				ctx: context.Background(),
				target: NopStrategyTarget{
					ReaderCloser: io.NopCloser(bytes.NewReader([]byte("docker"))),
				},
			},
			wantErr: false,
		},
		{
			name: "WithStartupTimeoutDefault sets context.Deadline for nil WaitStrategy.Timeout",
			strategy: ForAll(
				ForNop(
					func(ctx context.Context, _ StrategyTarget) error {
						if _, set := ctx.Deadline(); !set {
							return errors.New("expected context.Deadline to be set")
						}
						return nil
					},
				),
				ForLog("docker"),
			).WithStartupTimeoutDefault(1 * time.Second),
			args: args{
				ctx: context.Background(),
				target: NopStrategyTarget{
					ReaderCloser: io.NopCloser(bytes.NewReader([]byte("docker"))),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.strategy.WaitUntilReady(tt.args.ctx, tt.args.target)
			if tt.wantErr {
				require.Error(t, err, "ForAll.WaitUntilReady()")
			} else {
				require.NoErrorf(t, err, "ForAll.WaitUntilReady()")
			}
		})
	}
}

func TestMultiStrategy_handleNils(t *testing.T) {
	t.Run("nil-strategy", func(t *testing.T) {
		strategy := ForAll(nil)
		err := strategy.WaitUntilReady(context.Background(), NopStrategyTarget{})
		require.NoError(t, err)
	})

	t.Run("nil-strategy-in-the-middle", func(t *testing.T) {
		strategy := ForAll(nil, ForLog("docker"))
		err := strategy.WaitUntilReady(context.Background(), NopStrategyTarget{
			ReaderCloser: io.NopCloser(bytes.NewReader([]byte("docker"))),
		})
		require.NoError(t, err)
	})

	t.Run("nil-strategy-last", func(t *testing.T) {
		strategy := ForAll(ForLog("docker"), nil)
		err := strategy.WaitUntilReady(context.Background(), NopStrategyTarget{
			ReaderCloser: io.NopCloser(bytes.NewReader([]byte("docker"))),
		})
		require.NoError(t, err)
	})

	t.Run("nil-type-implements-strategy", func(t *testing.T) {
		var nilStrategy Strategy

		strategy := ForAll(ForLog("docker"), nilStrategy)
		err := strategy.WaitUntilReady(context.Background(), NopStrategyTarget{
			ReaderCloser: io.NopCloser(bytes.NewReader([]byte("docker"))),
		})
		require.NoError(t, err)
	})

	t.Run("nil-concrete-value-implements-strategy", func(t *testing.T) {
		// Create a nil pointer to a type that implements Strategy
		var nilPointerStrategy *nilWaitStrategy
		// When we assign it to the interface, the type information is preserved
		// but the concrete value is nil
		var strategyInterface Strategy = nilPointerStrategy

		strategy := ForAll(ForLog("docker"), strategyInterface)
		err := strategy.WaitUntilReady(context.Background(), NopStrategyTarget{
			ReaderCloser: io.NopCloser(bytes.NewReader([]byte("docker"))),
		})
		require.NoError(t, err)
	})
}

type nilWaitStrategy struct{}

func (s *nilWaitStrategy) WaitUntilReady(_ context.Context, _ StrategyTarget) error {
	return nil
}
