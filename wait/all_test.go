package wait

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"
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
					func(ctx context.Context, target StrategyTarget) error {
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
					func(ctx context.Context, target StrategyTarget) error {
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
					func(ctx context.Context, target StrategyTarget) error {
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
					func(ctx context.Context, target StrategyTarget) error {
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
			if err := tt.strategy.WaitUntilReady(tt.args.ctx, tt.args.target); (err != nil) != tt.wantErr {
				t.Errorf("ForAll.WaitUntilReady() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}
