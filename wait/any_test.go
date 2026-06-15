package wait

import (
	"bytes"
	"context"
	"errors"
	"io"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"
)

func TestAnyMultiStrategy_WaitsForAny(t *testing.T) {
	// synctest makes the time.Sleep below "instant".
	synctest.Test(t, func(t *testing.T) {
		var s1Done, s2Done atomic.Bool
		s1Release := make(chan struct{})
		s1 := ForNop(func(ctx context.Context, _ StrategyTarget) error {
			defer func() { s1Done.Store(true) }()

			// Releases only when we tell it to.
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-s1Release:
				return nil
			}
		})
		s2 := ForNop(func(ctx context.Context, _ StrategyTarget) error {
			defer func() { s2Done.Store(true) }()
			<-ctx.Done()
			return ctx.Err()
		})

		res := make(chan error)
		go func() {
			res <- ForAny(s1, s2).WaitUntilReady(t.Context(), NopStrategyTarget{})
		}()

		time.Sleep(time.Second)
		if s1Done.Load() || s2Done.Load() {
			t.Fatalf("no waiting should be done: s1=%v, s2=%v", s1Done.Load(), s2Done.Load())
		}

		close(s1Release)

		select {
		case <-t.Context().Done():
			t.Fatal(t.Context().Err())
		case err := <-res:
			if err != nil {
				t.Fatalf("expected no error, but got: %v", err)
			}
		}

		if !s1Done.Load() {
			t.Fatal("s1 should be done, but it is not")
		}
	})
}

func TestAnyMultiStrategy_FailuresNotPermitted(t *testing.T) {
	// When one strategy fails, we return its failure and stop.
	s1 := ForNop(func(ctx context.Context, _ StrategyTarget) error {
		<-ctx.Done()
		return ctx.Err()
	})
	s2 := ForNop(func(context.Context, StrategyTarget) error {
		return errors.New("s2 errored!")
	})

	res := make(chan error)
	go func() {
		res <- ForAny(s1, s2).WaitUntilReady(t.Context(), NopStrategyTarget{})
	}()

	select {
	case <-t.Context().Done():
		t.Fatal(t.Context().Err())
	case err := <-res:
		if err == nil {
			t.Fatalf("expected error, but got none: %v", err)
		}
	}
}

func TestAnyMultiStrategy_WaitUntilReady(t *testing.T) {
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
			strategy: ForAny(),
			args: args{
				ctx:    context.Background(),
				target: NopStrategyTarget{},
			},
			wantErr: true,
		},
		{
			name: "returns WaitStrategy error",
			strategy: ForAny(
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
			strategy: ForAny(
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
			strategy: ForAny(
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
			strategy: ForAny(
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
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.strategy.WaitUntilReady(tt.args.ctx, tt.args.target)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected err but there was none")
				}
			} else {
				if err != nil {
					t.Fatal("expected no err, but there was one")
				}
			}
		})
	}
}

func TestAnyMultiStrategy_handleNils(t *testing.T) {
	t.Run("nil-strategy", func(t *testing.T) {
		strategy := ForAny(nil)
		if err := strategy.WaitUntilReady(context.Background(), NopStrategyTarget{}); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("nil-strategy-in-the-middle", func(t *testing.T) {
		strategy := ForAny(nil, ForLog("docker"))
		if err := strategy.WaitUntilReady(context.Background(), NopStrategyTarget{
			ReaderCloser: io.NopCloser(bytes.NewReader([]byte("docker"))),
		}); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("nil-strategy-last", func(t *testing.T) {
		strategy := ForAny(ForLog("docker"), nil)
		if err := strategy.WaitUntilReady(context.Background(), NopStrategyTarget{
			ReaderCloser: io.NopCloser(bytes.NewReader([]byte("docker"))),
		}); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("nil-type-implements-strategy", func(t *testing.T) {
		var nilStrategy Strategy

		strategy := ForAny(ForLog("docker"), nilStrategy)
		if err := strategy.WaitUntilReady(context.Background(), NopStrategyTarget{
			ReaderCloser: io.NopCloser(bytes.NewReader([]byte("docker"))),
		}); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("nil-concrete-value-implements-strategy", func(t *testing.T) {
		// Create a nil pointer to a type that implements Strategy
		var nilPointerStrategy *nilWaitStrategy
		// When we assign it to the interface, the type information is preserved
		// but the concrete value is nil
		var strategyInterface Strategy = nilPointerStrategy

		strategy := ForAny(ForLog("docker"), strategyInterface)
		if err := strategy.WaitUntilReady(context.Background(), NopStrategyTarget{
			ReaderCloser: io.NopCloser(bytes.NewReader([]byte("docker"))),
		}); err != nil {
			t.Fatal(err)
		}
	})
}
