package sbx

import (
	"errors"
	"net"
	"net/http"
	"net/url"
	"time"
)

func newDockerTransport(endpoint *url.URL) (*http.Transport, error) {
	if endpoint == nil || endpoint.Host == "" {
		return nil, errors.New("initialize sbx Docker transport: endpoint is missing")
	}
	if endpoint.Scheme != "http" && endpoint.Scheme != "https" {
		return nil, errors.New("initialize sbx Docker transport: endpoint must use http or https")
	}
	return &http.Transport{
		Proxy:                 nil,
		DialContext:           (&net.Dialer{Timeout: 5 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
		ForceAttemptHTTP2:     false,
		MaxIdleConns:          32,
		MaxIdleConnsPerHost:   32,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: time.Second,
	}, nil
}
