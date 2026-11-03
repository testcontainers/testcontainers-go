package sbx

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"
)

const (
	diagnosticLimit = 8 * 1024
	stdoutLimit     = 1 * 1024 * 1024
)

type boundedOutput struct {
	data     []byte
	limit    int
	overflow bool
}

func (b *boundedOutput) Write(p []byte) (int, error) {
	remaining := b.limit - len(b.data)
	if len(p) > remaining {
		b.overflow = true
	}
	if remaining > 0 {
		if len(p) < remaining {
			remaining = len(p)
		}
		b.data = append(b.data, p[:remaining]...)
	}
	return len(p), nil
}

func newCommandRunner(cfg configuration) commandRunner {
	return func(ctx context.Context, args ...string) ([]byte, error) {
		all := make([]string, 0, len(args)+2)
		if cfg.appName != "" {
			all = append(all, "--app-name", cfg.appName)
		}
		all = append(all, args...)
		cmd := exec.CommandContext(ctx, cfg.binary, all...)
		cmd.WaitDelay = 2 * time.Second
		stdout := boundedOutput{limit: stdoutLimit}
		stderr := boundedOutput{limit: diagnosticLimit}
		cmd.Stdout, cmd.Stderr = &stdout, &stderr
		if err := cmd.Run(); err != nil {
			detail := strings.TrimSpace(string(stderr.data))
			if ctx.Err() != nil {
				if detail != "" {
					return nil, fmt.Errorf("sbx %s: %w: %s", commandLabel(args), ctx.Err(), detail)
				}
				return nil, fmt.Errorf("sbx %s: %w", commandLabel(args), ctx.Err())
			}
			if detail == "" {
				return nil, fmt.Errorf("sbx %s: %w", commandLabel(args), err)
			}
			return nil, fmt.Errorf("sbx %s: %w: %s", commandLabel(args), err, detail)
		}
		if stdout.overflow {
			return nil, fmt.Errorf("sbx %s output exceeded %d bytes", commandLabel(args), stdoutLimit)
		}
		return stdout.data, nil
	}
}

func commandLabel(args []string) string {
	if len(args) == 0 {
		return "command"
	}
	return args[0]
}

var _ io.Writer = (*boundedOutput)(nil)
