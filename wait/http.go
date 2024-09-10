package wait

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/docker/go-connections/nat"
)

// Implement interface
var (
	_ Strategy        = (*HTTPStrategy)(nil)
	_ StrategyTimeout = (*HTTPStrategy)(nil)
)

type HTTPStrategy struct {
	// all Strategies should have a startupTimeout to avoid waiting infinitely
	timeout *time.Duration

	// additional properties
	Port                   nat.Port
	Path                   string
	StatusCodeMatcher      func(status int) bool
	ResponseMatcher        func(body io.Reader) bool
	UseTLS                 bool
	AllowInsecure          bool
	TLSConfig              *tls.Config // TLS config for HTTPS
	Method                 string      // http method
	Body                   io.Reader   // http request body
	Headers                map[string]string
	ResponseHeadersMatcher func(headers http.Header) bool
	PollInterval           time.Duration
	UserInfo               *url.Userinfo
	ForceIPv4LocalHost     bool
}

// NewHTTPStrategy constructs a HTTP strategy waiting on port 80 and status code 200
func NewHTTPStrategy(path string) *HTTPStrategy {
	return &HTTPStrategy{
		Port:                   "",
		Path:                   path,
		StatusCodeMatcher:      defaultStatusCodeMatcher,
		ResponseMatcher:        func(body io.Reader) bool { return true },
		UseTLS:                 false,
		TLSConfig:              nil,
		Method:                 http.MethodGet,
		Body:                   nil,
		Headers:                map[string]string{},
		ResponseHeadersMatcher: func(headers http.Header) bool { return true },
		PollInterval:           defaultPollInterval(),
		UserInfo:               nil,
	}
}

func defaultStatusCodeMatcher(status int) bool {
	return status == http.StatusOK
}

// fluent builders for each property
// since go has neither covariance nor generics, the return type must be the type of the concrete implementation
// this is true for all properties, even the "shared" ones like startupTimeout

// WithStartupTimeout can be used to change the default startup timeout
func (ws *HTTPStrategy) WithStartupTimeout(timeout time.Duration) *HTTPStrategy {
	ws.timeout = &timeout
	return ws
}

// WithPort set the port to wait for.
// Default is the lowest numbered port.
func (ws *HTTPStrategy) WithPort(port nat.Port) *HTTPStrategy {
	ws.Port = port
	return ws
}

func (ws *HTTPStrategy) WithStatusCodeMatcher(statusCodeMatcher func(status int) bool) *HTTPStrategy {
	ws.StatusCodeMatcher = statusCodeMatcher
	return ws
}

func (ws *HTTPStrategy) WithResponseMatcher(matcher func(body io.Reader) bool) *HTTPStrategy {
	ws.ResponseMatcher = matcher
	return ws
}

func (ws *HTTPStrategy) WithTLS(useTLS bool, tlsconf ...*tls.Config) *HTTPStrategy {
	ws.UseTLS = useTLS
	if useTLS && len(tlsconf) > 0 {
		ws.TLSConfig = tlsconf[0]
	}
	return ws
}

func (ws *HTTPStrategy) WithAllowInsecure(allowInsecure bool) *HTTPStrategy {
	ws.AllowInsecure = allowInsecure
	return ws
}

func (ws *HTTPStrategy) WithMethod(method string) *HTTPStrategy {
	ws.Method = method
	return ws
}

func (ws *HTTPStrategy) WithBody(reqdata io.Reader) *HTTPStrategy {
	ws.Body = reqdata
	return ws
}

func (ws *HTTPStrategy) WithHeaders(headers map[string]string) *HTTPStrategy {
	ws.Headers = headers
	return ws
}

func (ws *HTTPStrategy) WithResponseHeadersMatcher(matcher func(http.Header) bool) *HTTPStrategy {
	ws.ResponseHeadersMatcher = matcher
	return ws
}

func (ws *HTTPStrategy) WithBasicAuth(username, password string) *HTTPStrategy {
	ws.UserInfo = url.UserPassword(username, password)
	return ws
}

// WithPollInterval can be used to override the default polling interval of 100 milliseconds
func (ws *HTTPStrategy) WithPollInterval(pollInterval time.Duration) *HTTPStrategy {
	ws.PollInterval = pollInterval
	return ws
}

// WithForcedIPv4LocalHost forces usage of localhost to be ipv4 127.0.0.1
// to avoid ipv6 docker bugs https://github.com/moby/moby/issues/42442 https://github.com/moby/moby/issues/42375
func (ws *HTTPStrategy) WithForcedIPv4LocalHost() *HTTPStrategy {
	ws.ForceIPv4LocalHost = true
	return ws
}

// ForHTTP is an alias for NewHTTPStrategy.
func ForHTTP(path string) *HTTPStrategy {
	return NewHTTPStrategy(path)
}

func (ws *HTTPStrategy) Timeout() *time.Duration {
	return ws.timeout
}

// WaitUntilReady implements Strategy.WaitUntilReady
func (ws *HTTPStrategy) WaitUntilReady(ctx context.Context, target StrategyTarget) error {
	timeout := defaultStartupTimeout()
	if ws.timeout != nil {
		timeout = *ws.timeout
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// TODO: Handle HTTP3 which is UDP.
	port, err := hostPortMapping(ctx, ws.Port, ws.PollInterval, ws.ForceIPv4LocalHost, "tcp", target)
	if err != nil {
		return err
	}

	if err := ws.httpCheck(ctx, target, port); err != nil {
		return err
	}

	return nil
}

// httpCheck sets up and runs the HTTP check until it succeeds or the context is canceled.
func (ws *HTTPStrategy) httpCheck(ctx context.Context, target StrategyTarget, port *portDetails) error {
	switch ws.Method {
	case http.MethodGet, http.MethodHead, http.MethodPost,
		http.MethodPut, http.MethodPatch, http.MethodDelete,
		http.MethodConnect, http.MethodOptions, http.MethodTrace:
	case "":
		ws.Method = http.MethodGet
	default:
		return fmt.Errorf("invalid http method %q", ws.Method)
	}

	tripper := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig:       ws.TLSConfig,
	}

	var proto string
	if ws.UseTLS {
		proto = "https"
		if ws.AllowInsecure {
			if ws.TLSConfig == nil {
				tripper.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
			} else {
				ws.TLSConfig.InsecureSkipVerify = true
			}
		}
	} else {
		proto = "http"
	}

	client := http.Client{Transport: tripper, Timeout: time.Second}

	endpoint, err := url.Parse(ws.Path)
	if err != nil {
		return err
	}
	endpoint.Scheme = proto
	endpoint.Host = port.Address()

	if ws.UserInfo != nil {
		endpoint.User = ws.UserInfo
	}

	// cache the body into a byte-slice so that it can be iterated over multiple times
	var body []byte
	if ws.Body != nil {
		body, err = io.ReadAll(ws.Body)
		if err != nil {
			return fmt.Errorf("read body: %w", err)
		}
	}

	matcher := ws.matcher()

	for {
		err := ws.doRequest(ctx, &client, endpoint, body, matcher)
		if err == nil {
			return nil
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("%w: %w", ctx.Err(), err)
		case <-time.After(ws.PollInterval):
			if err := checkTarget(ctx, target); err != nil {
				return err
			}
		}
	}
}

// warpMatcher wraps a matcher with a handler and a next matcher.
func warpMatcher(handler, next func(resp *http.Response) error) func(resp *http.Response) error {
	return func(resp *http.Response) error {
		if err := handler(resp); err != nil {
			return err
		}

		return next(resp)
	}
}

// matcher returns a matcher function that combines all the matchers.
func (ws *HTTPStrategy) matcher() func(resp *http.Response) error {
	matcher := func(resp *http.Response) error {
		return nil
	}
	if ws.StatusCodeMatcher == nil {
		matcher = warpMatcher(func(resp *http.Response) error {
			if !ws.StatusCodeMatcher(resp.StatusCode) {
				return fmt.Errorf("status code mismatch %d", resp.StatusCode)
			}
			return nil
		}, matcher)
	}
	if ws.ResponseMatcher == nil {
		matcher = warpMatcher(func(resp *http.Response) error {
			if !ws.ResponseMatcher(resp.Body) {
				return errors.New("body mismatch")
			}
			return nil
		}, matcher)
	}
	if ws.ResponseHeadersMatcher == nil {
		matcher = warpMatcher(func(resp *http.Response) error {
			if !ws.ResponseHeadersMatcher(resp.Header) {
				return fmt.Errorf("header mismatch: %v", resp.Header)
			}
			return nil
		}, matcher)
	}

	return matcher
}

// doRequest performs the actual HTTP request.
func (ws *HTTPStrategy) doRequest(ctx context.Context, client *http.Client, endpoint *url.URL, body []byte, matcher func(resp *http.Response) error) (err error) {
	req, err := http.NewRequestWithContext(ctx, ws.Method, endpoint.String(), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}

	for k, v := range ws.Headers {
		req.Header.Set(k, v)
	}

	var resp *http.Response
	if resp, err = client.Do(req); err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer func() {
		err = errors.Join(err, resp.Body.Close())
	}()

	if err := matcher(resp); err != nil {
		return err
	}

	return nil
}
