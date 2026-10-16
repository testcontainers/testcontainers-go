package sbx

import (
	"context"
	"net/url"
)

// commandRunner runs sbx with its configured binary and global options.
type commandRunner func(context.Context, ...string) ([]byte, error)

type sandboxRuntime struct {
	name      string
	run       commandRunner
	dockerURL *url.URL
	workspace string
}

type proxyServer interface {
	Host() string
	Close(context.Context) error
}
