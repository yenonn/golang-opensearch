package opensearch

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	opensearch "github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
)

// Client wraps the OpenSearch client with custom methods
type Client struct {
	client *opensearch.Client
}

// Config holds configuration for the OpenSearch client
type Config struct {
	Addresses []string
	Username  string
	Password  string
	// InsecureSkipVerify skips TLS certificate verification (use for development only)
	InsecureSkipVerify bool
}

// NewClient creates a new OpenSearch client with the provided configuration
func NewClient(config Config) (*Client, error) {
	if len(config.Addresses) == 0 {
		return nil, fmt.Errorf("at least one address is required")
	}

	cfg := opensearch.Config{
		Addresses: config.Addresses,
		Username:  config.Username,
		Password:  config.Password,
	}

	// Configure TLS if needed
	if config.InsecureSkipVerify {
		cfg.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}

	client, err := opensearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenSearch client: %w", err)
	}

	return &Client{client: client}, nil
}

// Ping checks if the OpenSearch cluster is reachable
func (c *Client) Ping(ctx context.Context) error {
	req := opensearchapi.PingRequest{}
	res, err := req.Do(ctx, c.client)
	if err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("ping failed with status: %s", res.Status())
	}

	return nil
}

// Info returns information about the OpenSearch cluster
func (c *Client) Info(ctx context.Context) (map[string]interface{}, error) {
	res, err := c.client.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster info: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("info request failed with status: %s", res.Status())
	}

	var response map[string]interface{}
	if err := parseResponse(res.Body, &response); err != nil {
		return nil, err
	}

	return response, nil
}

// GetClient returns the underlying OpenSearch client for advanced usage
func (c *Client) GetClient() *opensearch.Client {
	return c.client
}