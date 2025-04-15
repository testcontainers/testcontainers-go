package client

// openAIEndpointSuffix is the suffix for the OpenAI endpoint
const openAIEndpointSuffix = "/engines/v1"

// Client is the client for the Docker Model Runner
type Client struct {
	baseURL        string
	openAIEndpoint string
}

// NewClient creates a new client for the Docker Model Runner
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL:        baseURL,
		openAIEndpoint: baseURL + openAIEndpointSuffix,
	}
}

// OpenAIEndpoint returns the OpenAI endpoint for the Docker Model Runner
func (c *Client) OpenAIEndpoint() string {
	return c.openAIEndpoint
}
