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

// AdminAPIClient is a client for the Redpanda Admin API.
type AdminAPIClient struct {
	BaseURL  string
	username string
	password string
	client   *http.Client
}

// NewAdminAPIClient creates a new AdminAPIClient.
func NewAdminAPIClient(baseURL string) *AdminAPIClient {
	return &AdminAPIClient{
		BaseURL: baseURL,
		client:  http.DefaultClient,
	}
}

// WithHTTPClient sets the HTTP client for the AdminAPIClient.
func (cl *AdminAPIClient) WithHTTPClient(c *http.Client) *AdminAPIClient {
	cl.client = c
	return cl
}

// WithAuthentication sets the username and password for the AdminAPIClient.
func (cl *AdminAPIClient) WithAuthentication(username, password string) *AdminAPIClient {
	cl.username = username
	cl.password = password
	return cl
}

// Username returns the username of the AdminAPIClient.
func (cl *AdminAPIClient) Username() string {
	return cl.username
}

// Password returns the password of the AdminAPIClient.
func (cl *AdminAPIClient) Password() string {
	return cl.password
}

type createUserRequest struct {
	User      string `json:"username,omitempty"`
	Password  string `json:"password"`
	Algorithm string `json:"algorithm"`
}

// CreateUser creates a new user in Redpanda using Admin API.
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

	if cl.username != "" || cl.password != "" {
		req.SetBasicAuth(cl.username, cl.password)
	}

	resp, err := cl.client.Do(req)
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
