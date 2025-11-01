# Claude Code Documentation

This document provides comprehensive information about the Golang OpenSearch Client project for Claude Code and AI assistants.

## Project Overview

**Project Name:** Golang OpenSearch Client
**Language:** Go 1.25.3
**Purpose:** A reusable Go client library for interacting with OpenSearch running on Minikube, providing clean CRUD operations and query builders.

## Architecture

### Directory Structure

```
golang-opensearch/
├── main.go                          # Demo application showcasing all CRUD operations
├── pkg/opensearch/                  # Reusable library code
│   ├── client.go                    # Client initialization, connection management
│   ├── crud.go                      # CRUD operations (Create, Read, Update, Delete)
│   └── models.go                    # Response types, query builders, helpers
├── build/                           # Kubernetes manifests
│   ├── opensearch-deployment.yaml   # OpenSearch deployment
│   └── opensearch-dashboard.yaml    # OpenSearch Dashboards deployment
├── bin/                             # Build output (gitignored)
├── go.mod                           # Go module definition
├── go.sum                           # Dependency checksums
├── Makefile                         # Build and deployment automation
├── README.md                        # User-facing documentation
└── claude.md                        # This file (AI assistant documentation)
```

### Key Components

#### 1. Client Library (`pkg/opensearch/`)

**client.go** - Connection and client management
- `Config` struct: Configuration for OpenSearch connection
  - `Addresses []string`: List of OpenSearch endpoints
  - `Username string`: Optional basic auth username
  - `Password string`: Optional basic auth password
  - `InsecureSkipVerify bool`: Skip TLS verification (dev only)
- `Client` struct: Wraps opensearch-go client
- `NewClient(config Config) (*Client, error)`: Initialize client
- `Ping(ctx context.Context) error`: Health check
- `Info(ctx context.Context) (map[string]interface{}, error)`: Cluster info
- `GetClient() *opensearch.Client`: Access underlying client

**crud.go** - Document and index operations
- Document Operations:
  - `CreateDocument(ctx, index, id string, document interface{}) error`
  - `GetDocument(ctx, index, id string) (map[string]interface{}, error)`
  - `UpdateDocument(ctx, index, id string, updates interface{}) error`
  - `DeleteDocument(ctx, index, id string) error`
  - `BulkCreate(ctx, index string, documents []map[string]interface{}) error`
- Search Operations:
  - `SearchDocuments(ctx, index string, query map[string]interface{}) ([]map[string]interface{}, error)`
  - `SearchAll(ctx, index string) ([]map[string]interface{}, error)`
- Index Operations:
  - `CreateIndex(ctx, index string, body map[string]interface{}) error`
  - `DeleteIndex(ctx, index string) error`
  - `IndexExists(ctx, index string) (bool, error)`

**models.go** - Types and query builders
- Response Types:
  - `GetResponse`: Document retrieval response
  - `SearchResponse`: Search query response
  - `Hit`: Individual search result
  - `BulkResponse`: Bulk operation response
  - `BulkItem`: Single item in bulk response
  - `ErrorResponse`: Error response structure
- Query Builders:
  - `MatchAllQuery() map[string]interface{}`
  - `MatchQuery(field, value string) map[string]interface{}`
  - `TermQuery(field string, value interface{}) map[string]interface{}`
  - `RangeQuery(field string, gte, lte interface{}) map[string]interface{}`
  - `BoolQuery(must, should, mustNot []map[string]interface{}) map[string]interface{}`
- Query Modifiers:
  - `WithSize(query map[string]interface{}, size int)`
  - `WithFrom(query map[string]interface{}, from int)`
  - `WithSort(query map[string]interface{}, field, order string)`

#### 2. Demo Application (`main.go`)

Comprehensive example demonstrating all library features:
1. Connection setup and testing
2. Index creation and management
3. Single document creation
4. Bulk document creation
5. Document retrieval
6. Search operations (match, range queries)
7. Document updates
8. Document deletion
9. Cleanup operations

#### 3. Infrastructure

**Makefile Targets:**

Cluster Management:
- `make start`: Start Minikube cluster and deploy OpenSearch
- `make stop`: Stop OpenSearch and Minikube
- `make delete`: Delete entire cluster
- `make status`: Check cluster status
- `make deploy-opensearch`: Deploy/redeploy OpenSearch only
- `make deploy-dashboard`: Deploy/redeploy dashboards only

Application:
- `make run`: Run the Go application
- `make build`: Build binary to bin/opensearch-app
- `make test`: Run tests
- `make clean`: Remove build artifacts

Utilities:
- `make port-forward`: Set up kubectl port forwarding (for Docker driver)
- `make tunnel`: Create minikube service tunnel (alternative)
- `make get-minikube-ip`: Display Minikube IP

## Development Guidelines

### Connection Configuration

The application supports two connection modes:

**1. Port Forwarding (Default - for Docker driver):**
```bash
# Terminal 1
make port-forward  # Forwards to localhost:9200

# Terminal 2
make run  # Uses OPENSEARCH_URL=http://localhost:9200
```

**2. Direct NodePort Access (Linux or other drivers):**
```bash
OPENSEARCH_URL=http://$(minikube ip):30920 make run
```

### Environment Variables

- `OPENSEARCH_URL`: OpenSearch endpoint (default: `http://localhost:9200`)
  - For port-forward: `http://localhost:9200`
  - For NodePort: `http://<minikube-ip>:30920`

### Code Patterns

#### Creating a Client

```go
config := opensearch.Config{
    Addresses: []string{"http://localhost:9200"},
    Username:  "admin",  // Optional
    Password:  "admin",  // Optional
}

client, err := opensearch.NewClient(config)
if err != nil {
    log.Fatal(err)
}
```

#### CRUD Operations

```go
ctx := context.Background()

// Create
doc := map[string]interface{}{
    "title": "Example",
    "value": 123,
}
err := client.CreateDocument(ctx, "my-index", "doc-1", doc)

// Read
doc, err := client.GetDocument(ctx, "my-index", "doc-1")

// Update
updates := map[string]interface{}{"value": 456}
err := client.UpdateDocument(ctx, "my-index", "doc-1", updates)

// Delete
err := client.DeleteDocument(ctx, "my-index", "doc-1")
```

#### Search Operations

```go
// Search all
results, err := client.SearchAll(ctx, "my-index")

// Match query
query := opensearch.MatchQuery("title", "example")
results, err := client.SearchDocuments(ctx, "my-index", query)

// Range query
query := opensearch.RangeQuery("value", 100, 500)
results, err := client.SearchDocuments(ctx, "my-index", query)

// Complex query
query := opensearch.BoolQuery(
    []map[string]interface{}{
        opensearch.MatchQuery("category", "tutorial"),
    },
    []map[string]interface{}{
        opensearch.RangeQuery("views", 100, nil),
    },
    nil,
)
query = opensearch.WithSize(query, 10)
query = opensearch.WithSort(query, "views", "desc")
results, err := client.SearchDocuments(ctx, "my-index", query)
```

#### Bulk Operations

```go
docs := []map[string]interface{}{
    {
        "_id": "doc-1",
        "title": "Document 1",
        "value": 100,
    },
    {
        "_id": "doc-2",
        "title": "Document 2",
        "value": 200,
    },
}

err := client.BulkCreate(ctx, "my-index", docs)
```

## Common Issues and Solutions

### Issue: "no route to host" error

**Cause:** Using Minikube with Docker driver on macOS/Windows
**Solution:** Use port forwarding
```bash
make port-forward  # Keep running in separate terminal
```

### Issue: Connection refused on localhost:9200

**Cause:** Port forwarding not running
**Solution:** Ensure `make port-forward` is active in another terminal

### Issue: Document not found after creation

**Cause:** Index refresh not complete
**Solution:** All CRUD operations use `refresh=true` parameter, but for manual operations add a small delay or check index refresh settings

## Testing

### Manual Testing
```bash
# Start cluster
make start

# Setup port forwarding
make port-forward

# In new terminal, run application
make run
```

### Expected Output
The demo application should output:
- Connection test results
- Index creation confirmation
- Document creation confirmations (single and bulk)
- Search results with match counts
- Update confirmations
- Delete confirmations
- Final document count and summary

## Dependencies

**External:**
- `github.com/opensearch-project/opensearch-go/v2`: Official OpenSearch Go client

**Development:**
- Go 1.25.3+
- Minikube
- kubectl

## Future Enhancements

Potential improvements for this project:
1. Add unit tests for library functions
2. Add integration tests with test containers
3. Implement connection pooling configuration
4. Add retry logic for transient failures
5. Support for SSL/TLS connections
6. Add more query builders (geo, nested, aggregations)
7. Implement document streaming for large result sets
8. Add metrics and logging middleware
9. Create examples directory with more use cases
10. Add CI/CD pipeline configuration

## Contributing Guidelines

When modifying this project:
1. Keep library code (`pkg/opensearch/`) separate from application code
2. All public functions should have clear error handling
3. Use context.Context for all operations
4. Follow Go naming conventions
5. Update README.md when adding new features
6. Test with both port-forward and direct access methods
7. Ensure compatibility with OpenSearch 2.x API

## References

- [OpenSearch Go Client Documentation](https://github.com/opensearch-project/opensearch-go)
- [OpenSearch REST API Reference](https://opensearch.org/docs/latest/api-reference/)
- [Minikube Documentation](https://minikube.sigs.k8s.io/docs/)

---

**Last Updated:** 2025-11-01
**Version:** 1.0.0
**Branch:** lib-dev-1