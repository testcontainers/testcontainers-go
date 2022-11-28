package exec

import (
	"bytes"
	"io"

	"github.com/docker/docker/pkg/stdcopy"
)

// ProcessOptions defines options applicable to the reader processor
type ProcessOptions struct {
	Reader io.Reader
}

// ProcessOption defines a common interface to modify the reader processor
// These options can be passed to the Exec function in a variadic way to customize the returned Reader instance
type ProcessOption interface {
	Apply(opts *ProcessOptions)
}

type ProcessOptionFunc func(opts *ProcessOptions)

func (fn ProcessOptionFunc) Apply(opts *ProcessOptions) {
	fn(opts)
}

func Multiplexed() ProcessOption {
	return ProcessOptionFunc(func(opts *ProcessOptions) {
		done := make(chan struct{})

		var outBuff bytes.Buffer
		var errBuff bytes.Buffer
		go func() {
			if _, err := stdcopy.StdCopy(&outBuff, &errBuff, opts.Reader); err != nil {
				return
			}
			close(done)
		}()

		<-done

		opts.Reader = &outBuff
	})
}
