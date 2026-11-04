// Package testrest provides bounded JSON HTTP transport for emulator tests.
package testrest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Do sends one JSON request. The caller supplies a client timeout and context
// deadline; service requests and response assertions remain in the tests.
func Do(ctx context.Context, client *http.Client, method, endpoint string, body, result any) error {
	var payload []byte
	var err error
	if body != nil {
		payload, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encode %s %s: %w", method, endpoint, err)
		}
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("request %s %s: %w", method, endpoint, err)
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("send %s %s: %w", method, endpoint, err)
	}
	defer resp.Body.Close()
	const maxBody = 1 << 20
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxBody+1))
	if err != nil {
		return fmt.Errorf("read %s %s: %w", method, endpoint, err)
	}
	if len(data) > maxBody {
		return fmt.Errorf("%s %s: response exceeds %d bytes", method, endpoint, maxBody)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("%s %s: %s: %s", method, endpoint, resp.Status, data)
	}
	if result != nil {
		if err := json.Unmarshal(data, result); err != nil {
			return fmt.Errorf("decode %s %s: %w; body: %s", method, endpoint, err, data)
		}
	}
	return nil
}
