package opensearch

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"
)

// TestClient is a helper to create a client for integration tests
func setupTestClient(t *testing.T) *Client {
	t.Helper()

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
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test connection
	ctx := context.Background()
	if err := client.Ping(ctx); err != nil {
		t.Skipf("OpenSearch not available at %s: %v", url, err)
	}

	return client
}

// setupTestIndex creates a test index and returns cleanup function
func setupTestIndex(t *testing.T, client *Client, indexName string) func() {
	t.Helper()
	ctx := context.Background()

	// Delete index if it exists
	exists, _ := client.IndexExists(ctx, indexName)
	if exists {
		_ = client.DeleteIndex(ctx, indexName)
	}

	// Create fresh index
	err := client.CreateIndex(ctx, indexName, nil)
	if err != nil {
		t.Fatalf("Failed to create test index: %v", err)
	}

	// Wait for index to be ready
	time.Sleep(100 * time.Millisecond)

	return func() {
		_ = client.DeleteIndex(ctx, indexName)
	}
}

func TestCreateDocument(t *testing.T) {
	client := setupTestClient(t)
	indexName := "test-create-doc"
	cleanup := setupTestIndex(t, client, indexName)
	defer cleanup()

	ctx := context.Background()

	tests := []struct {
		name      string
		docID     string
		document  interface{}
		wantError bool
	}{
		{
			name:  "Create simple document",
			docID: "doc-1",
			document: map[string]interface{}{
				"title": "Test Document",
				"value": 123,
			},
			wantError: false,
		},
		{
			name:  "Create document with nested fields",
			docID: "doc-2",
			document: map[string]interface{}{
				"title": "Nested Document",
				"metadata": map[string]interface{}{
					"author": "Test Author",
					"tags":   []string{"test", "example"},
				},
			},
			wantError: false,
		},
		{
			name:  "Create document with array",
			docID: "doc-3",
			document: map[string]interface{}{
				"items": []int{1, 2, 3, 4, 5},
			},
			wantError: false,
		},
		{
			name:  "Overwrite existing document",
			docID: "doc-1",
			document: map[string]interface{}{
				"title": "Updated Document",
				"value": 456,
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.CreateDocument(ctx, indexName, tt.docID, tt.document)
			if (err != nil) != tt.wantError {
				t.Errorf("CreateDocument() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestGetDocument(t *testing.T) {
	client := setupTestClient(t)
	indexName := "test-get-doc"
	cleanup := setupTestIndex(t, client, indexName)
	defer cleanup()

	ctx := context.Background()

	// Create test documents
	testDoc := map[string]interface{}{
		"title": "Test Document",
		"value": 123,
		"tags":  []string{"test", "example"},
	}
	err := client.CreateDocument(ctx, indexName, "existing-doc", testDoc)
	if err != nil {
		t.Fatalf("Failed to create test document: %v", err)
	}

	tests := []struct {
		name      string
		docID     string
		wantError bool
		validate  func(t *testing.T, doc map[string]interface{})
	}{
		{
			name:      "Get existing document",
			docID:     "existing-doc",
			wantError: false,
			validate: func(t *testing.T, doc map[string]interface{}) {
				if doc["title"] != "Test Document" {
					t.Errorf("Expected title 'Test Document', got %v", doc["title"])
				}
				if doc["value"] != float64(123) { // JSON numbers are float64
					t.Errorf("Expected value 123, got %v", doc["value"])
				}
			},
		},
		{
			name:      "Get non-existent document",
			docID:     "non-existent",
			wantError: true,
			validate:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := client.GetDocument(ctx, indexName, tt.docID)
			if (err != nil) != tt.wantError {
				t.Errorf("GetDocument() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && tt.validate != nil {
				tt.validate(t, doc)
			}
		})
	}
}

func TestUpdateDocument(t *testing.T) {
	client := setupTestClient(t)
	indexName := "test-update-doc"
	cleanup := setupTestIndex(t, client, indexName)
	defer cleanup()

	ctx := context.Background()

	// Create test document
	testDoc := map[string]interface{}{
		"title":       "Original Title",
		"value":       100,
		"description": "Original description",
	}
	err := client.CreateDocument(ctx, indexName, "doc-1", testDoc)
	if err != nil {
		t.Fatalf("Failed to create test document: %v", err)
	}

	tests := []struct {
		name      string
		docID     string
		updates   interface{}
		wantError bool
		validate  func(t *testing.T, doc map[string]interface{})
	}{
		{
			name:  "Update single field",
			docID: "doc-1",
			updates: map[string]interface{}{
				"value": 200,
			},
			wantError: false,
			validate: func(t *testing.T, doc map[string]interface{}) {
				if doc["value"] != float64(200) {
					t.Errorf("Expected value 200, got %v", doc["value"])
				}
				if doc["title"] != "Original Title" {
					t.Errorf("Title should remain unchanged, got %v", doc["title"])
				}
			},
		},
		{
			name:  "Update multiple fields",
			docID: "doc-1",
			updates: map[string]interface{}{
				"title":       "Updated Title",
				"description": "Updated description",
			},
			wantError: false,
			validate: func(t *testing.T, doc map[string]interface{}) {
				if doc["title"] != "Updated Title" {
					t.Errorf("Expected title 'Updated Title', got %v", doc["title"])
				}
				if doc["description"] != "Updated description" {
					t.Errorf("Expected updated description, got %v", doc["description"])
				}
			},
		},
		{
			name:  "Update non-existent document",
			docID: "non-existent",
			updates: map[string]interface{}{
				"value": 300,
			},
			wantError: true,
			validate:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.UpdateDocument(ctx, indexName, tt.docID, tt.updates)
			if (err != nil) != tt.wantError {
				t.Errorf("UpdateDocument() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError && tt.validate != nil {
				// Verify update
				doc, err := client.GetDocument(ctx, indexName, tt.docID)
				if err != nil {
					t.Fatalf("Failed to get updated document: %v", err)
				}
				tt.validate(t, doc)
			}
		})
	}
}

func TestDeleteDocument(t *testing.T) {
	client := setupTestClient(t)
	indexName := "test-delete-doc"
	cleanup := setupTestIndex(t, client, indexName)
	defer cleanup()

	ctx := context.Background()

	// Create test document
	testDoc := map[string]interface{}{
		"title": "Document to Delete",
	}
	err := client.CreateDocument(ctx, indexName, "doc-to-delete", testDoc)
	if err != nil {
		t.Fatalf("Failed to create test document: %v", err)
	}

	tests := []struct {
		name      string
		docID     string
		wantError bool
	}{
		{
			name:      "Delete existing document",
			docID:     "doc-to-delete",
			wantError: false,
		},
		{
			name:      "Delete non-existent document",
			docID:     "non-existent",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.DeleteDocument(ctx, indexName, tt.docID)
			if (err != nil) != tt.wantError {
				t.Errorf("DeleteDocument() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError {
				// Verify deletion
				_, err := client.GetDocument(ctx, indexName, tt.docID)
				if err == nil {
					t.Error("Document should not exist after deletion")
				}
			}
		})
	}
}

func TestSearchDocuments(t *testing.T) {
	client := setupTestClient(t)
	indexName := "test-search-docs"
	cleanup := setupTestIndex(t, client, indexName)
	defer cleanup()

	ctx := context.Background()

	// Create test documents
	testDocs := []struct {
		id  string
		doc map[string]interface{}
	}{
		{
			id: "doc-1",
			doc: map[string]interface{}{
				"title":    "Golang Tutorial",
				"category": "programming",
				"views":    150,
			},
		},
		{
			id: "doc-2",
			doc: map[string]interface{}{
				"title":    "Python Guide",
				"category": "programming",
				"views":    200,
			},
		},
		{
			id: "doc-3",
			doc: map[string]interface{}{
				"title":    "Cooking Recipe",
				"category": "food",
				"views":    50,
			},
		},
	}

	for _, td := range testDocs {
		err := client.CreateDocument(ctx, indexName, td.id, td.doc)
		if err != nil {
			t.Fatalf("Failed to create test document %s: %v", td.id, err)
		}
	}

	// Wait for documents to be indexed
	time.Sleep(200 * time.Millisecond)

	tests := []struct {
		name           string
		query          map[string]interface{}
		wantError      bool
		wantMinResults int
		wantMaxResults int
		validate       func(t *testing.T, results []map[string]interface{})
	}{
		{
			name: "Match all query",
			query: map[string]interface{}{
				"query": map[string]interface{}{
					"match_all": map[string]interface{}{},
				},
			},
			wantError:      false,
			wantMinResults: 3,
			wantMaxResults: 3,
		},
		{
			name: "Match query on title",
			query: map[string]interface{}{
				"query": map[string]interface{}{
					"match": map[string]interface{}{
						"title": "tutorial",
					},
				},
			},
			wantError:      false,
			wantMinResults: 1,
			wantMaxResults: 1,
			validate: func(t *testing.T, results []map[string]interface{}) {
				if len(results) > 0 {
					title := results[0]["title"].(string)
					if title != "Golang Tutorial" {
						t.Errorf("Expected 'Golang Tutorial', got %v", title)
					}
				}
			},
		},
		{
			name: "Term query on category",
			query: map[string]interface{}{
				"query": map[string]interface{}{
					"term": map[string]interface{}{
						"category": "programming",
					},
				},
			},
			wantError:      false,
			wantMinResults: 2,
			wantMaxResults: 2,
		},
		{
			name: "Range query on views",
			query: map[string]interface{}{
				"query": map[string]interface{}{
					"range": map[string]interface{}{
						"views": map[string]interface{}{
							"gte": 100,
							"lte": 250,
						},
					},
				},
			},
			wantError:      false,
			wantMinResults: 2,
			wantMaxResults: 2,
		},
		{
			name: "Query with size limit",
			query: map[string]interface{}{
				"query": map[string]interface{}{
					"match_all": map[string]interface{}{},
				},
				"size": 2,
			},
			wantError:      false,
			wantMinResults: 2,
			wantMaxResults: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := client.SearchDocuments(ctx, indexName, tt.query)
			if (err != nil) != tt.wantError {
				t.Errorf("SearchDocuments() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError {
				if len(results) < tt.wantMinResults || len(results) > tt.wantMaxResults {
					t.Errorf("Expected %d-%d results, got %d", tt.wantMinResults, tt.wantMaxResults, len(results))
				}

				// Verify results have _id and _score
				for _, result := range results {
					if _, ok := result["_id"]; !ok {
						t.Error("Result should have _id field")
					}
					if _, ok := result["_score"]; !ok {
						t.Error("Result should have _score field")
					}
				}

				if tt.validate != nil {
					tt.validate(t, results)
				}
			}
		})
	}
}

func TestSearchAll(t *testing.T) {
	client := setupTestClient(t)
	indexName := "test-search-all"
	cleanup := setupTestIndex(t, client, indexName)
	defer cleanup()

	ctx := context.Background()

	// Create multiple test documents
	for i := 1; i <= 5; i++ {
		doc := map[string]interface{}{
			"id":    i,
			"title": fmt.Sprintf("Document %d", i),
		}
		err := client.CreateDocument(ctx, indexName, fmt.Sprintf("doc-%d", i), doc)
		if err != nil {
			t.Fatalf("Failed to create test document: %v", err)
		}
	}

	// Wait for documents to be indexed
	time.Sleep(200 * time.Millisecond)

	results, err := client.SearchAll(ctx, indexName)
	if err != nil {
		t.Fatalf("SearchAll() error = %v", err)
	}

	if len(results) != 5 {
		t.Errorf("Expected 5 results, got %d", len(results))
	}

	// Verify all results have required fields
	for _, result := range results {
		if _, ok := result["_id"]; !ok {
			t.Error("Result should have _id field")
		}
		if _, ok := result["_score"]; !ok {
			t.Error("Result should have _score field")
		}
		if _, ok := result["title"]; !ok {
			t.Error("Result should have title field")
		}
	}
}

func TestCreateIndex(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	tests := []struct {
		name      string
		indexName string
		body      map[string]interface{}
		wantError bool
	}{
		{
			name:      "Create index without settings",
			indexName: "test-index-simple",
			body:      nil,
			wantError: false,
		},
		{
			name:      "Create index with settings",
			indexName: "test-index-settings",
			body: map[string]interface{}{
				"settings": map[string]interface{}{
					"number_of_shards":   1,
					"number_of_replicas": 0,
				},
			},
			wantError: false,
		},
		{
			name:      "Create index with mappings",
			indexName: "test-index-mappings",
			body: map[string]interface{}{
				"mappings": map[string]interface{}{
					"properties": map[string]interface{}{
						"title": map[string]interface{}{
							"type": "text",
						},
						"timestamp": map[string]interface{}{
							"type": "date",
						},
					},
				},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Cleanup before test
			exists, _ := client.IndexExists(ctx, tt.indexName)
			if exists {
				_ = client.DeleteIndex(ctx, tt.indexName)
			}

			err := client.CreateIndex(ctx, tt.indexName, tt.body)
			if (err != nil) != tt.wantError {
				t.Errorf("CreateIndex() error = %v, wantError %v", err, tt.wantError)
			}

			// Cleanup after test
			if !tt.wantError {
				_ = client.DeleteIndex(ctx, tt.indexName)
			}
		})
	}
}

func TestDeleteIndex(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	tests := []struct {
		name      string
		setup     func(t *testing.T) string
		wantError bool
	}{
		{
			name: "Delete existing index",
			setup: func(t *testing.T) string {
				indexName := "test-index-to-delete"
				err := client.CreateIndex(ctx, indexName, nil)
				if err != nil {
					t.Fatalf("Failed to create test index: %v", err)
				}
				return indexName
			},
			wantError: false,
		},
		{
			name: "Delete non-existent index",
			setup: func(t *testing.T) string {
				return "non-existent-index"
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			indexName := tt.setup(t)

			err := client.DeleteIndex(ctx, indexName)
			if (err != nil) != tt.wantError {
				t.Errorf("DeleteIndex() error = %v, wantError %v", err, tt.wantError)
			}

			if !tt.wantError {
				// Verify deletion
				exists, _ := client.IndexExists(ctx, indexName)
				if exists {
					t.Error("Index should not exist after deletion")
				}
			}
		})
	}
}

func TestIndexExists(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	// Create a test index
	existingIndex := "test-index-exists"
	err := client.CreateIndex(ctx, existingIndex, nil)
	if err != nil {
		t.Fatalf("Failed to create test index: %v", err)
	}
	defer client.DeleteIndex(ctx, existingIndex)

	tests := []struct {
		name       string
		indexName  string
		wantExists bool
		wantError  bool
	}{
		{
			name:       "Check existing index",
			indexName:  existingIndex,
			wantExists: true,
			wantError:  false,
		},
		{
			name:       "Check non-existent index",
			indexName:  "non-existent-index",
			wantExists: false,
			wantError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists, err := client.IndexExists(ctx, tt.indexName)
			if (err != nil) != tt.wantError {
				t.Errorf("IndexExists() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if exists != tt.wantExists {
				t.Errorf("IndexExists() = %v, want %v", exists, tt.wantExists)
			}
		})
	}
}

func TestBulkCreate(t *testing.T) {
	client := setupTestClient(t)
	indexName := "test-bulk-create"
	cleanup := setupTestIndex(t, client, indexName)
	defer cleanup()

	ctx := context.Background()

	tests := []struct {
		name         string
		documents    []map[string]interface{}
		wantError    bool
		validateFunc func(t *testing.T)
	}{
		{
			name:      "Bulk create empty list",
			documents: []map[string]interface{}{},
			wantError: false,
		},
		{
			name: "Bulk create with IDs",
			documents: []map[string]interface{}{
				{
					"_id":   "bulk-1",
					"title": "Bulk Document 1",
					"value": 100,
				},
				{
					"_id":   "bulk-2",
					"title": "Bulk Document 2",
					"value": 200,
				},
				{
					"_id":   "bulk-3",
					"title": "Bulk Document 3",
					"value": 300,
				},
			},
			wantError: false,
			validateFunc: func(t *testing.T) {
				// Wait for documents to be indexed
				time.Sleep(200 * time.Millisecond)

				// Verify documents were created
				for i := 1; i <= 3; i++ {
					docID := fmt.Sprintf("bulk-%d", i)
					doc, err := client.GetDocument(ctx, indexName, docID)
					if err != nil {
						t.Errorf("Failed to get document %s: %v", docID, err)
					}
					if doc["value"] != float64(i*100) {
						t.Errorf("Document %s has incorrect value: %v", docID, doc["value"])
					}
				}
			},
		},
		{
			name: "Bulk create without IDs",
			documents: []map[string]interface{}{
				{
					"title": "Auto ID Document 1",
					"value": 10,
				},
				{
					"title": "Auto ID Document 2",
					"value": 20,
				},
			},
			wantError: false,
			validateFunc: func(t *testing.T) {
				// Wait for documents to be indexed
				time.Sleep(200 * time.Millisecond)

				// Verify documents exist via search
				results, err := client.SearchAll(ctx, indexName)
				if err != nil {
					t.Fatalf("Failed to search documents: %v", err)
				}
				if len(results) < 2 {
					t.Errorf("Expected at least 2 documents, got %d", len(results))
				}
			},
		},
		{
			name: "Bulk create large batch",
			documents: func() []map[string]interface{} {
				docs := make([]map[string]interface{}, 100)
				for i := 0; i < 100; i++ {
					docs[i] = map[string]interface{}{
						"_id":   fmt.Sprintf("large-batch-%d", i),
						"index": i,
						"data":  fmt.Sprintf("Data %d", i),
					}
				}
				return docs
			}(),
			wantError: false,
			validateFunc: func(t *testing.T) {
				// Wait for documents to be indexed
				time.Sleep(500 * time.Millisecond)

				// Verify some documents were created
				doc, err := client.GetDocument(ctx, indexName, "large-batch-50")
				if err != nil {
					t.Errorf("Failed to get document from large batch: %v", err)
				}
				if doc["index"] != float64(50) {
					t.Errorf("Document has incorrect index: %v", doc["index"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.BulkCreate(ctx, indexName, tt.documents)
			if (err != nil) != tt.wantError {
				t.Errorf("BulkCreate() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError && tt.validateFunc != nil {
				tt.validateFunc(t)
			}
		})
	}
}

// TestIntegrationWorkflow tests a complete CRUD workflow
func TestIntegrationWorkflow(t *testing.T) {
	client := setupTestClient(t)
	indexName := "test-integration"
	cleanup := setupTestIndex(t, client, indexName)
	defer cleanup()

	ctx := context.Background()

	// 1. Create documents
	t.Log("Step 1: Creating documents")
	docs := []map[string]interface{}{
		{"_id": "user-1", "name": "Alice", "age": 30, "role": "admin"},
		{"_id": "user-2", "name": "Bob", "age": 25, "role": "user"},
		{"_id": "user-3", "name": "Charlie", "age": 35, "role": "user"},
	}
	err := client.BulkCreate(ctx, indexName, docs)
	if err != nil {
		t.Fatalf("Failed to create documents: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	// 2. Search all
	t.Log("Step 2: Searching all documents")
	results, err := client.SearchAll(ctx, indexName)
	if err != nil {
		t.Fatalf("Failed to search all: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("Expected 3 documents, got %d", len(results))
	}

	// 3. Get specific document
	t.Log("Step 3: Getting specific document")
	user, err := client.GetDocument(ctx, indexName, "user-1")
	if err != nil {
		t.Fatalf("Failed to get document: %v", err)
	}
	if user["name"] != "Alice" {
		t.Errorf("Expected name 'Alice', got %v", user["name"])
	}

	// 4. Update document
	t.Log("Step 4: Updating document")
	err = client.UpdateDocument(ctx, indexName, "user-1", map[string]interface{}{
		"age": 31,
	})
	if err != nil {
		t.Fatalf("Failed to update document: %v", err)
	}

	// Verify update
	user, err = client.GetDocument(ctx, indexName, "user-1")
	if err != nil {
		t.Fatalf("Failed to get updated document: %v", err)
	}
	if user["age"] != float64(31) {
		t.Errorf("Expected age 31, got %v", user["age"])
	}

	// 5. Search with query
	t.Log("Step 5: Searching with query")
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"term": map[string]interface{}{
				"role": "user",
			},
		},
	}
	results, err = client.SearchDocuments(ctx, indexName, query)
	if err != nil {
		t.Fatalf("Failed to search with query: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 user role documents, got %d", len(results))
	}

	// 6. Delete document
	t.Log("Step 6: Deleting document")
	err = client.DeleteDocument(ctx, indexName, "user-2")
	if err != nil {
		t.Fatalf("Failed to delete document: %v", err)
	}

	// Verify deletion
	_, err = client.GetDocument(ctx, indexName, "user-2")
	if err == nil {
		t.Error("Document should not exist after deletion")
	}

	// Final count
	results, err = client.SearchAll(ctx, indexName)
	if err != nil {
		t.Fatalf("Failed to search all: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 remaining documents, got %d", len(results))
	}

	t.Log("Integration workflow completed successfully")
}

// Helper function to pretty print JSON for debugging
func prettyPrint(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("%+v", v)
	}
	return string(b)
}