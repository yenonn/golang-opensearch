# Golang OpenSearch Client

A Go client library and application for interacting with OpenSearch running on Minikube. This project provides a clean, reusable library for standard CRUD operations on OpenSearch.

## Overview

This project demonstrates how to:
- Connect to OpenSearch running in a Minikube cluster
- Perform CRUD (Create, Read, Update, Delete) operations
- Organize OpenSearch client code as a reusable library
- Manage OpenSearch deployment with Makefile commands

## Technologies

- **Go 1.25.3**: Programming language
- **OpenSearch**: Search and analytics engine
- **Minikube**: Local Kubernetes cluster
- **opensearch-go**: Official OpenSearch Go client

## Prerequisites

Before you begin, ensure you have the following installed:

- [Go](https://golang.org/dl/) (version 1.25.3 or later)
- [Minikube](https://minikube.sigs.k8s.io/docs/start/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)

## Quick Start

### 1. Start OpenSearch on Minikube

```bash
make start
```

This command will:
- Start a Minikube cluster with profile `my-elasticsearch-cluster`
- Deploy OpenSearch and OpenSearch Dashboards
- Expose OpenSearch API on port 30920
- Expose OpenSearch Dashboards on port 30561

### 2. Set Up Port Forwarding (Required for Docker Driver)

If you're using Minikube with the Docker driver (common on macOS/Windows), you need to set up port forwarding:

```bash
make port-forward
```

Keep this terminal open. OpenSearch will be available at `http://localhost:9200`.

**Alternative:** Use `make tunnel` for minikube service tunnel (also requires keeping terminal open).

### 3. Run the Application

In a **new terminal**, run:

```bash
make run
```

This will execute the Go application and demonstrate CRUD operations.

### 3. Stop the Cluster

```bash
make stop
```

## Project Structure

```
golang-opensearch/
├── main.go                    # Application entry point with examples
├── pkg/
│   └── opensearch/
│       ├── client.go          # Client initialization & connection
│       ├── crud.go            # CRUD operations
│       └── models.go          # Common types & response structures
├── build/
│   ├── opensearch-deployment.yaml
│   └── opensearch-dashboard.yaml
├── go.mod                     # Go module file
├── go.sum                     # Dependency checksums
├── Makefile                   # Build and deployment commands
└── README.md                  # This file
```

## Configuration

### OpenSearch Connection

The application connects to OpenSearch using the Minikube IP and NodePort service (port 30920).

To get your Minikube IP:
```bash
minikube ip --profile my-elasticsearch-cluster
```

The default endpoint is: `http://<minikube-ip>:30920`

## CRUD Operations

The library provides the following operations:

### Create (Index a Document)

```go
err := client.CreateDocument(ctx, "my-index", "doc-id", document)
```

### Read (Get a Document)

```go
doc, err := client.GetDocument(ctx, "my-index", "doc-id")
```

### Search Documents

```go
results, err := client.SearchDocuments(ctx, "my-index", query)
```

### Update a Document

```go
err := client.UpdateDocument(ctx, "my-index", "doc-id", updates)
```

### Delete a Document

```go
err := client.DeleteDocument(ctx, "my-index", "doc-id")
```

## Makefile Commands

### Cluster Management

- `make start` - Start Minikube cluster and deploy OpenSearch
- `make stop` - Stop OpenSearch deployments and Minikube cluster
- `make delete` - Delete the Minikube cluster entirely
- `make status` - Check cluster status

### OpenSearch Deployment

- `make deploy-opensearch` - Deploy/redeploy OpenSearch only
- `make deploy-dashboard` - Deploy/redeploy OpenSearch Dashboards only

### Application

- `make run` - Run the Go application
- `make build` - Build the application binary

### Help

- `make help` - Display available commands

## Development

### Installing Dependencies

```bash
go mod download
```

### Building the Project

```bash
go build -o bin/opensearch-app .
```

### Running the Application

```bash
go run main.go
```

## API Reference

### Client Initialization

```go
import "github.com/yenonn/go-opensearch/pkg/opensearch"

config := opensearch.Config{
    Addresses: []string{"http://<minikube-ip>:30920"},
}

client, err := opensearch.NewClient(config)
```

### Available Methods

- `NewClient(config Config) (*Client, error)` - Create new OpenSearch client
- `CreateDocument(ctx context.Context, index, id string, document interface{}) error`
- `GetDocument(ctx context.Context, index, id string) (map[string]interface{}, error)`
- `SearchDocuments(ctx context.Context, index string, query map[string]interface{}) ([]map[string]interface{}, error)`
- `UpdateDocument(ctx context.Context, index, id string, updates interface{}) error`
- `DeleteDocument(ctx context.Context, index, id string) error`
- `Ping(ctx context.Context) error` - Health check

## Troubleshooting

### Cannot Connect to OpenSearch - "no route to host"

This error typically occurs when using Minikube with the Docker driver on macOS/Windows. The Docker driver doesn't allow direct NodePort access.

**Solution:** Use port forwarding or minikube tunnel:

```bash
# Option 1: kubectl port-forward (Recommended)
make port-forward
# Then in another terminal:
make run

# Option 2: minikube service tunnel
make tunnel
# Update OPENSEARCH_URL environment variable to use the tunneled URL
```

### For Direct NodePort Access (Linux or other drivers)

If you're using a driver that supports direct NodePort access:

```bash
# Get the Minikube IP
MINIKUBE_IP=$(minikube ip --profile my-elasticsearch-cluster)

# Run with custom URL
OPENSEARCH_URL=http://$MINIKUBE_IP:30920 make run
```

### General Connection Issues

1. Check if Minikube is running:
   ```bash
   make status
   ```

2. Verify OpenSearch pod is ready:
   ```bash
   kubectl --context my-elasticsearch-cluster get pods
   ```

3. Test OpenSearch directly (with port-forward active):
   ```bash
   curl http://localhost:9200
   ```

### Port Already in Use

If ports 30920 or 30561 are already in use, modify the NodePort values in:
- `build/opensearch-deployment.yaml`
- `build/opensearch-dashboard.yaml`

### OpenSearch Not Starting

Check pod logs:
```bash
kubectl --context my-elasticsearch-cluster logs deployment/opensearch
```

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.