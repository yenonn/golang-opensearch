package opensearch

import (
	"context"
	"os"
	"testing"

	opensearch "github.com/opensearch-project/opensearch-go/v2"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		wantError bool
	# github.com/yenonn/go-opensearch
./main.go:130:2: fmt.Println arg list ends with redundant newline
FAIL    github.com/yenonn/go-opensearch [build failed]
	errorMsg  string
	}{
		{
			name: "Valid config with single address",
			config: Config{
				Addresses:          []string{"http://localhost:9200"},
				Username:           "admin",
				Password:           "admin",
				InsecureSkipVerify: true,
			},
			wantError: false,
		},
		{
			name: "Valid config with multiple addresses",
			config: Config{
				Addresses:          []string{"http://localhost:9200", "http://localhost:9201"},
				Username:           "admin",
				Password:           "admin",
				InsecureSkipVerify: true,
			},
			wantError: false,
		},
		{
			name: "Valid config without auth",
			config: Config{
				Addresses:          []string{"http://localhost:9200"},
				InsecureSkipVerify: true,
			},
			wantError: false,
		},
		{
			name: "Valid config with TLS verification enabled",
			config: Config{
				Addresses:          []string{"https://localhost:9200"},
				Username:           "admin",
				Password:           "admin",
				InsecureSkipVerify: false,
			},
			wantError: false,
		},
		{
			name:      "Empty addresses - should fail",
			config:    Config{},
			wantError: true,
			errorMsg:  "at least one address is required",
		},
		{
			name: "Empty addresses slice - should fail",
			config: Config{
				Addresses: []string{},
			},
			wantError: true,
			errorMsg:  "at least one address is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.config)

			if tt.wantError {
				if err == nil {
					t.Errorf("NewClient() expected error but got nil")
					return
				}
				if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("NewClient() error = %v, want error containing %v", err.Error(), tt.errorMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("NewClient() unexpected error = %v", err)
				return
			}

			if client == nil {
				t.Error("NewClient() returned nil client")
				return
			}

			// Verify client has underlying OpenSearch client
			if client.client == nil {
				t.Error("NewClient() client.client is nil")
			}
		})
	}
}

func TestNewClient_ConfigValidation(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		validate func(t *testing.T, client *Client)
	}{
		{
			name: "Config with InsecureSkipVerify creates proper transport",
			config: Config{
				Addresses:          []string{"https://localhost:9200"},
				InsecureSkipVerify: true,
			},
			validate: func(t *testing.T, client *Client) {
				if client == nil || client.client == nil {
					t.Fatal("Client should not be nil")
				}
				// Client is created successfully with TLS config
			},
		},
		{
			name: "Config with credentials",
			config: Config{
				Addresses: []string{"http://localhost:9200"},
				Username:  "testuser",
				Password:  "testpass",
			},
			validate: func(t *testing.T, client *Client) {
				if client == nil || client.client == nil {
					t.Fatal("Client should not be nil")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.config)
			if err != nil {
				t.Fatalf("NewClient() error = %v", err)
			}
			if tt.validate != nil {
				tt.validate(t, client)
			}
		})
	}
}

func TestClient_Ping(t *testing.T) {
	url := os.Getenv("OPENSEARCH_URL")
	if url == "" {
		url = "http://localhost:9200"
	}

	tests := []struct {
		name      string
		config    Config
		wantError bool
		skipMsg   string
	}{
		{
			name: "Ping valid OpenSearch cluster",
			config: Config{
				Addresses:          []string{url},
				Username:           "admin",
				Password:           "admin",
				InsecureSkipVerify: true,
			},
			wantError: false,
			skipMsg:   "requires OpenSearch to be running",
		},
		{
			name: "Ping invalid address",
			config: Config{
				Addresses:          []string{"http://invalid-host:9999"},
				InsecureSkipVerify: true,
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.config)
			if err != nil {
				t.Fatalf("NewClient() error = %v", err)
			}

			ctx := context.Background()
			err = client.Ping(ctx)

			if tt.wantError {
				if err == nil {
					t.Error("Ping() expected error but got nil")
				}
				return
			}

			// For valid config, skip if OpenSearch is not available
			if err != nil && tt.skipMsg != "" {
				t.Skipf("Skipping test: %s - %v", tt.skipMsg, err)
				return
			}

			if err != nil {
				t.Errorf("Ping() unexpected error = %v", err)
			}
		})
	}
}

func TestClient_Ping_WithContext(t *testing.T) {
	url := os.Getenv("OPENSEARCH_URL")
	if url == "" {
		url = "http://localhost:9200"
	}

	config := Config{
		Addresses:          []string{url},
		Username:           "admin",
		Password:           "admin",
		InsecureSkipVerify: true,
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	ctx := context.Background()
	err = client.Ping(ctx)
	if err != nil {
		t.Skipf("OpenSearch not available: %v", err)
	}

	// Test with background context
	t.Run("Ping with background context", func(t *testing.T) {
		ctx := context.Background()
		err := client.Ping(ctx)
		if err != nil {
			t.Errorf("Ping() with background context error = %v", err)
		}
	})

	// Test with TODO context
	t.Run("Ping with TODO context", func(t *testing.T) {
		ctx := context.TODO()
		err := client.Ping(ctx)
		if err != nil {
			t.Errorf("Ping() with TODO context error = %v", err)
		}
	})
}

func TestClient_Info(t *testing.T) {
	url := os.Getenv("OPENSEARCH_URL")
	if url == "" {
		url = "http://localhost:9200"
	}

	tests := []struct {
		name      string
		config    Config
		wantError bool
		validate  func(t *testing.T, info map[string]interface{})
		skipMsg   string
	}{
		{
			name: "Get cluster info from valid cluster",
			config: Config{
				Addresses:          []string{url},
				Username:           "admin",
				Password:           "admin",
				InsecureSkipVerify: true,
			},
			wantError: false,
			validate: func(t *testing.T, info map[string]interface{}) {
				// Check for standard OpenSearch/Elasticsearch info fields
				if info == nil {
					t.Error("Info() returned nil map")
					return
				}

				// Check for common fields in cluster info
				expectedFields := []string{"name", "cluster_name", "version"}
				for _, field := range expectedFields {
					if _, ok := info[field]; !ok {
						t.Logf("Warning: Info response missing field: %s", field)
						t.Logf("Full response: %+v", info)
					}
				}

				// Verify version info exists and has structure
				if version, ok := info["version"]; ok {
					if versionMap, ok := version.(map[string]interface{}); ok {
						if _, ok := versionMap["number"]; !ok {
							t.Error("Version info missing 'number' field")
						}
					}
				}
			},
			skipMsg: "requires OpenSearch to be running",
		},
		{
			name: "Get info from invalid address",
			config: Config{
				Addresses:          []string{"http://invalid-host:9999"},
				InsecureSkipVerify: true,
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.config)
			if err != nil {
				t.Fatalf("NewClient() error = %v", err)
			}

			ctx := context.Background()
			info, err := client.Info(ctx)

			if tt.wantError {
				if err == nil {
					t.Error("Info() expected error but got nil")
				}
				return
			}

			// For valid config, skip if OpenSearch is not available
			if err != nil && tt.skipMsg != "" {
				t.Skipf("Skipping test: %s - %v", tt.skipMsg, err)
				return
			}

			if err != nil {
				t.Errorf("Info() unexpected error = %v", err)
				return
			}

			if tt.validate != nil {
				tt.validate(t, info)
			}
		})
	}
}

func TestClient_GetClient(t *testing.T) {
	config := Config{
		Addresses:          []string{"http://localhost:9200"},
		Username:           "admin",
		Password:           "admin",
		InsecureSkipVerify: true,
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	t.Run("GetClient returns underlying client", func(t *testing.T) {
		underlyingClient := client.GetClient()

		if underlyingClient == nil {
			t.Error("GetClient() returned nil")
			return
		}

		// Verify it's an actual OpenSearch client
		var _ *opensearch.Client = underlyingClient
	})

	t.Run("GetClient returns same instance", func(t *testing.T) {
		client1 := client.GetClient()
		client2 := client.GetClient()

		if client1 != client2 {
			t.Error("GetClient() should return the same client instance")
		}
	})
}

func TestClient_Integration(t *testing.T) {
	url := os.Getenv("OPENSEARCH_URL")
	if url == "" {
		url = "http://localhost:9200"
	}

	config := Config{
		Addresses:          []string{url},
		Username:           "admin",
		Password:           "admin",
		InsecureSkipVerify: true,
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	ctx := context.Background()

	// Test Ping
	t.Run("Integration: Ping", func(t *testing.T) {
		err := client.Ping(ctx)
		if err != nil {
			t.Skipf("OpenSearch not available: %v", err)
		}
	})

	// Test Info
	t.Run("Integration: Info after Ping", func(t *testing.T) {
		// First ping to ensure connection
		err := client.Ping(ctx)
		if err != nil {
			t.Skipf("OpenSearch not available: %v", err)
		}

		// Then get info
		info, err := client.Info(ctx)
		if err != nil {
			t.Errorf("Info() error = %v", err)
			return
		}

		if len(info) == 0 {
			t.Error("Info() returned empty map")
		}

		t.Logf("Cluster info: %+v", info)
	})

	// Test GetClient and use it directly
	t.Run("Integration: Use underlying client", func(t *testing.T) {
		underlyingClient := client.GetClient()
		if underlyingClient == nil {
			t.Fatal("GetClient() returned nil")
		}

		// Use underlying client to call Info
		res, err := underlyingClient.Info()
		if err != nil {
			t.Skipf("OpenSearch not available: %v", err)
		}
		defer res.Body.Close()

		if res.IsError() {
			t.Errorf("Underlying client Info() failed with status: %s", res.Status())
		}
	})
}

// TestClient_Concurrent tests thread safety of client operations
func TestClient_Concurrent(t *testing.T) {
	url := os.Getenv("OPENSEARCH_URL")
	if url == "" {
		url = "http://localhost:9200"
	}

	config := Config{
		Addresses:          []string{url},
		Username:           "admin",
		Password:           "admin",
		InsecureSkipVerify: true,
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	ctx := context.Background()

	// Check if OpenSearch is available
	err = client.Ping(ctx)
	if err != nil {
		t.Skipf("OpenSearch not available: %v", err)
	}

	t.Run("Concurrent Ping operations", func(t *testing.T) {
		const numGoroutines = 10
		errChan := make(chan error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func() {
				errChan <- client.Ping(ctx)
			}()
		}

		// Collect results
		for i := 0; i < numGoroutines; i++ {
			if err := <-errChan; err != nil {
				t.Errorf("Concurrent Ping() error = %v", err)
			}
		}
	})

	t.Run("Concurrent Info operations", func(t *testing.T) {
		const numGoroutines = 10
		type result struct {
			info map[string]interface{}
			err  error
		}
		resultChan := make(chan result, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func() {
				info, err := client.Info(ctx)
				resultChan <- result{info: info, err: err}
			}()
		}

		// Collect results
		for i := 0; i < numGoroutines; i++ {
			res := <-resultChan
			if res.err != nil {
				t.Errorf("Concurrent Info() error = %v", res.err)
			}
			if res.info == nil {
				t.Error("Concurrent Info() returned nil map")
			}
		}
	})
}

// TestConfig_EdgeCases tests edge cases in configuration
func TestConfig_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		wantError bool
	}{
		{
			name: "Config with whitespace in address",
			config: Config{
				Addresses: []string{"  http://localhost:9200  "},
			},
			wantError: false, // Client creation succeeds, connection may fail
		},
		{
			name: "Config with empty username but has password",
			config: Config{
				Addresses: []string{"http://localhost:9200"},
				Username:  "",
				Password:  "password",
			},
			wantError: false,
		},
		{
			name: "Config with username but empty password",
			config: Config{
				Addresses: []string{"http://localhost:9200"},
				Username:  "admin",
				Password:  "",
			},
			wantError: false,
		},
		{
			name: "Config with HTTPS and InsecureSkipVerify",
			config: Config{
				Addresses:          []string{"https://localhost:9200"},
				InsecureSkipVerify: true,
			},
			wantError: false,
		},
		{
			name: "Config with HTTP (no S) and InsecureSkipVerify false",
			config: Config{
				Addresses:          []string{"http://localhost:9200"},
				InsecureSkipVerify: false,
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.config)

			if tt.wantError {
				if err == nil {
					t.Error("NewClient() expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("NewClient() unexpected error = %v", err)
				return
			}

			if client == nil {
				t.Error("NewClient() returned nil client")
			}
		})
	}
}
