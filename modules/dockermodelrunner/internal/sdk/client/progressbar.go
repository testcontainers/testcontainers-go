package client

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/schollz/progressbar/v3"
)

// ProgressWriter is an interface for progress reporting
type ProgressWriter interface {
	io.Writer
	SetTotal(total int64)
}

// ProgressBarAdapter adapts progressbar.ProgressBar to our ProgressWriter interface
type ProgressBarAdapter struct {
	*progressbar.ProgressBar
}

// SetTotal sets the total number of bytes to download
func (p *ProgressBarAdapter) SetTotal(total int64) {
	p.ChangeMax64(total)
}

// WithProgress sets a progress writer for the pull operation
func WithProgress(w ProgressWriter) PullOption {
	return func(opts *pullOptions) error {
		opts.progress = w
		return nil
	}
}

// WithProgressBar sets a progress writer for the pull operation, using the provided
// writer.
func WithProgressBar(w io.Writer, we io.Writer, total int) PullOption {
	return func(opts *pullOptions) error {
		if w == nil {
			return errors.New("writer is nil")
		}
		if we == nil {
			return errors.New("error writer is nil")
		}

		opts.progress = NewProgressBar(w, we, total)
		return nil
	}
}

// WithStdoutProgressBar sets a progress writer for the pull operation, using stdout
// as the writer.
func WithStdoutProgressBar(total int) PullOption {
	return func(opts *pullOptions) error {
		opts.progress = NewProgressBar(os.Stdout, os.Stderr, total)
		return nil
	}
}

// NewProgressBar creates a new progress bar
func NewProgressBar(w io.Writer, we io.Writer, total int) *ProgressBarAdapter {
	bar := progressbar.NewOptions(total,
		progressbar.OptionSetWriter(w),
		progressbar.OptionSetDescription("Pulling model..."),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(10),
		progressbar.OptionShowCount(),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprint(we, "\n")
		}),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionThrottle(100*time.Millisecond),
	)

	return &ProgressBarAdapter{ProgressBar: bar}
}
