package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/yenonn/go-opensearch/pkg/opensearch"
)

func main() {
	// Get OpenSearch address from environment variable or use default
	// When using kubectl port-forward, use: OPENSEARCH_URL=http://localhost:9200
	// When using direct NodePort access, use: OPENSEARCH_URL=http://<minikube-ip>:30920
	opensearchURL := os.Getenv("OPENSEARCH_URL")
	if opensearchURL == "" {
		opensearchURL = "http://localhost:9200" // Default for port-forward
		log.Printf("OPENSEARCH_URL not set, using default: %s", opensearchURL)
		log.Printf("Make sure to run 'make port-forward' in another terminal if using Docker driver")
	}

	// Create OpenSearch client configuration
	config := opensearch.Config{
		Addresses: []string{opensearchURL},
		// Username:  "admin",  // Uncomment if authentication is required
		// Password:  "admin",  // Uncomment if authentication is required
	}

	// Initialize the client
	client, err := opensearch.NewClient(config)
	if err != nil {
		log.Fatalf("Failed to create OpenSearch client: %v", err)
	}

	ctx := context.Background()

	// Test connection
	fmt.Println("=== Testing Connection ===")
	if err := client.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping OpenSearch: %v", err)
	}
	fmt.Println("✓ Successfully connected to OpenSearch")

	// Get cluster info
	info, err := client.Info(ctx)
	if err != nil {
		log.Fatalf("Failed to get cluster info: %v", err)
	}
	fmt.Printf("✓ Cluster info: %v\n\n", info["version"])

	// Define index name
	indexName := "demo-index"

	// Clean up: Delete index if it exists
	fmt.Println("=== Setting Up ===")
	exists, err := client.IndexExists(ctx, indexName)
	if err != nil {
		log.Printf("Warning: Failed to check index existence: %v", err)
	}
	if exists {
		if err := client.DeleteIndex(ctx, indexName); err != nil {
			log.Printf("Warning: Failed to delete existing index: %v", err)
		} else {
			fmt.Printf("✓ Deleted existing index: %s\n", indexName)
		}
	}

	// Create index
	if err := client.CreateIndex(ctx, indexName, nil); err != nil {
		log.Fatalf("Failed to create index: %v", err)
	}
	fmt.Printf("✓ Created index: %s\n\n", indexName)

	// === CREATE Operations ===
	fmt.Println("=== CREATE Operations ===")

	// Create single document
	doc1 := map[string]interface{}{
		"title":      "Introduction to OpenSearch",
		"author":     "John Doe",
		"category":   "tutorial",
		"views":      150,
		"published":  true,
		"created_at": time.Now().Format(time.RFC3339),
	}
	if err := client.CreateDocument(ctx, indexName, "1", doc1); err != nil {
		log.Fatalf("Failed to create document: %v", err)
	}
	fmt.Println("✓ Created document with ID: 1")

	doc2 := map[string]interface{}{
		"title":      "Advanced OpenSearch Queries",
		"author":     "Jane Smith",
		"category":   "advanced",
		"views":      320,
		"published":  true,
		"created_at": time.Now().Format(time.RFC3339),
	}
	if err := client.CreateDocument(ctx, indexName, "2", doc2); err != nil {
		log.Fatalf("Failed to create document: %v", err)
	}
	fmt.Println("✓ Created document with ID: 2")

	// Bulk create documents
	bulkDocs := []map[string]interface{}{
		{
			"_id":        "3",
			"title":      "Getting Started with Go",
			"author":     "Bob Johnson",
			"category":   "tutorial",
			"views":      200,
			"published":  true,
			"created_at": time.Now().Format(time.RFC3339),
		},
		{
			"_id":        "4",
			"title":      "OpenSearch Performance Tips",
			"author":     "Alice Williams",
			"category":   "advanced",
			"views":      450,
			"published":  false,
			"created_at": time.Now().Format(time.RFC3339),
		},
	}
	if err := client.BulkCreate(ctx, indexName, bulkDocs); err != nil {
		log.Fatalf("Failed to bulk create documents: %v", err)
	}
	fmt.Println("✓ Bulk created 2 documents\n")

	// Small delay to ensure indexing is complete
	time.Sleep(1 * time.Second)

	// === READ Operations ===
	fmt.Println("=== READ Operations ===")

	// Get single document
	doc, err := client.GetDocument(ctx, indexName, "1")
	if err != nil {
		log.Fatalf("Failed to get document: %v", err)
	}
	fmt.Printf("✓ Retrieved document ID 1: %s\n", doc["title"])

	// Search all documents
	allDocs, err := client.SearchAll(ctx, indexName)
	if err != nil {
		log.Fatalf("Failed to search all documents: %v", err)
	}
	fmt.Printf("✓ Found %d documents in index\n", len(allDocs))

	// Search with match query
	query := opensearch.MatchQuery("category", "tutorial")
	results, err := client.SearchDocuments(ctx, indexName, query)
	if err != nil {
		log.Fatalf("Failed to search documents: %v", err)
	}
	fmt.Printf("✓ Found %d tutorial documents\n", len(results))
	for _, result := range results {
		fmt.Printf("  - %s (by %s)\n", result["title"], result["author"])
	}

	// Search with NOT match query (exclude documents)
	notQuery := opensearch.NotMatchQuery("category", "tutorial")
	notResults, err := client.SearchDocuments(ctx, indexName, notQuery)
	if err != nil {
		log.Fatalf("Failed to search with NOT match query: %v", err)
	}
	fmt.Printf("✓ Found %d documents that are NOT tutorials\n", len(notResults))
	for _, result := range notResults {
		fmt.Printf("  - %s (category: %s)\n", result["title"], result["category"])
	}

	// Search with NOT term query (exclude exact match)
	notTermQuery := opensearch.NotTermQuery("published", false)
	notTermResults, err := client.SearchDocuments(ctx, indexName, notTermQuery)
	if err != nil {
		log.Fatalf("Failed to search with NOT term query: %v", err)
	}
	fmt.Printf("✓ Found %d published documents (excluding unpublished)\n", len(notTermResults))
	for _, result := range notTermResults {
		fmt.Printf("  - %s (published: %v)\n", result["title"], result["published"])
	}

	// Search with MatchMapQuery (match multiple fields)
	matchMapQuery := opensearch.MatchMapQuery(map[string]interface{}{
		"category": "tutorial",
		"author":   "John Doe",
	})
	matchMapResults, err := client.SearchDocuments(ctx, indexName, matchMapQuery)
	if err != nil {
		log.Fatalf("Failed to search with MatchMapQuery: %v", err)
	}
	fmt.Printf("✓ Found %d documents matching both category='tutorial' AND author='John Doe'\n", len(matchMapResults))
	for _, result := range matchMapResults {
		fmt.Printf("  - %s (author: %s, category: %s)\n", result["title"], result["author"], result["category"])
	}

	// Search with NotMatchMapQuery (exclude documents matching multiple fields)
	notMatchMapQuery := opensearch.NotMatchMapQuery(map[string]interface{}{
		"category": "tutorial",
		"author":   "Jane Smith",
	})
	notMatchMapResults, err := client.SearchDocuments(ctx, indexName, notMatchMapQuery)
	if err != nil {
		log.Fatalf("Failed to search with NotMatchMapQuery: %v", err)
	}
	fmt.Printf("✓ Found %d documents NOT matching category='tutorial' AND NOT matching author='Jane Smith'\n", len(notMatchMapResults))
	for _, result := range notMatchMapResults {
		fmt.Printf("  - %s (author: %s, category: %s)\n", result["title"], result["author"], result["category"])
	}

	// Search with range query
	rangeQuery := opensearch.RangeQuery("views", 200, 400)
	rangeResults, err := client.SearchDocuments(ctx, indexName, rangeQuery)
	if err != nil {
		log.Fatalf("Failed to search with range query: %v", err)
	}
	fmt.Printf("✓ Found %d documents with 200-400 views\n", len(rangeResults))
	fmt.Println()

	// === UPDATE Operations ===
	fmt.Println("=== UPDATE Operations ===")

	updates := map[string]interface{}{
		"views":      500,
		"updated_at": time.Now().Format(time.RFC3339),
	}
	if err := client.UpdateDocument(ctx, indexName, "1", updates); err != nil {
		log.Fatalf("Failed to update document: %v", err)
	}
	fmt.Println("✓ Updated document ID 1 (increased views to 500)")

	// Verify update
	updatedDoc, err := client.GetDocument(ctx, indexName, "1")
	if err != nil {
		log.Fatalf("Failed to get updated document: %v", err)
	}
	fmt.Printf("✓ Verified: Document ID 1 now has %.0f views\n\n", updatedDoc["views"])

	// === DELETE Operations ===
	fmt.Println("=== DELETE Operations ===")

	if err := client.DeleteDocument(ctx, indexName, "4"); err != nil {
		log.Fatalf("Failed to delete document: %v", err)
	}
	fmt.Println("✓ Deleted document ID 4")

	// Verify deletion
	allDocsAfterDelete, err := client.SearchAll(ctx, indexName)
	if err != nil {
		log.Fatalf("Failed to search all documents: %v", err)
	}
	fmt.Printf("✓ Verified: Now %d documents remain in index\n\n", len(allDocsAfterDelete))

	// === Summary ===
	fmt.Println("=== Summary ===")
	fmt.Println("✓ All CRUD operations completed successfully!")
	fmt.Printf("✓ Final document count: %d\n", len(allDocsAfterDelete))
	fmt.Println("\nRemaining documents:")
	for _, doc := range allDocsAfterDelete {
		fmt.Printf("  - ID %s: %s (%.0f views)\n", doc["_id"], doc["title"], doc["views"])
	}

	// Optional: Clean up the demo index
	fmt.Println("\n=== Cleanup ===")
	if err := client.DeleteIndex(ctx, indexName); err != nil {
		log.Printf("Warning: Failed to delete index: %v", err)
	} else {
		fmt.Printf("✓ Deleted demo index: %s\n", indexName)
	}

	fmt.Println("\n✓ Demo completed successfully!")
}
