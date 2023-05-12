package redpanda

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type AdminAPIClient struct {
	BaseURL string
}

func NewAdminAPIClient(baseURL string) *AdminAPIClient {
	return &AdminAPIClient{BaseURL: baseURL}
}

type createUserRequest struct {
	User      string `json:"username,omitempty"`
	Password  string `json:"password"`
	Algorithm string `json:"algorithm"`
}

func (cl *AdminAPIClient) CreateUser(ctx context.Context, username, password string) error {
	userReq := createUserRequest{
		User:      username,
		Password:  password,
		Algorithm: "SCRAM-SHA-256",
	}
	jsonReq, err := json.Marshal(userReq)
	if err != nil {
		return fmt.Errorf("failed to marshal create user request: %w", err)
	}

	endpoint, err := url.JoinPath(cl.BaseURL, "/v1/security/users")
	if err != nil {
		return fmt.Errorf("failed to join url path: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(jsonReq))
	if err != nil {
		return fmt.Errorf("failed to build http request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		return fmt.Errorf("unexpected status code in response: %d. Response body is: %q", resp.StatusCode, body)
	}

	return nil
}
