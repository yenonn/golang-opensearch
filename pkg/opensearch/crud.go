package opensearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
)

// CreateDocument indexes a new document or updates an existing one
func (c *Client) CreateDocument(ctx context.Context, index, id string, document interface{}) error {
	body, err := json.Marshal(document)
	if err != nil {
		return fmt.Errorf("failed to marshal document: %w", err)
	}

	req := opensearchapi.IndexRequest{
		Index:      index,
		DocumentID: id,
		Body:       bytes.NewReader(body),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, c.client)
	if err != nil {
		return fmt.Errorf("failed to index document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("index request failed with status: %s", res.Status())
	}

	return nil
}

// GetDocument retrieves a document by its ID
func (c *Client) GetDocument(ctx context.Context, index, id string) (map[string]interface{}, error) {
	req := opensearchapi.GetRequest{
		Index:      index,
		DocumentID: id,
	}

	res, err := req.Do(ctx, c.client)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == 404 {
			return nil, fmt.Errorf("document not found")
		}
		return nil, fmt.Errorf("get request failed with status: %s", res.Status())
	}

	var response GetResponse
	if err := parseResponse(res.Body, &response); err != nil {
		return nil, err
	}

	return response.Source, nil
}

// UpdateDocument updates an existing document with partial updates
func (c *Client) UpdateDocument(ctx context.Context, index, id string, updates interface{}) error {
	updateDoc := map[string]interface{}{
		"doc": updates,
	}

	body, err := json.Marshal(updateDoc)
	if err != nil {
		return fmt.Errorf("failed to marshal updates: %w", err)
	}

	req := opensearchapi.UpdateRequest{
		Index:      index,
		DocumentID: id,
		Body:       bytes.NewReader(body),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, c.client)
	if err != nil {
		return fmt.Errorf("failed to update document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == 404 {
			return fmt.Errorf("document not found")
		}
		return fmt.Errorf("update request failed with status: %s", res.Status())
	}

	return nil
}

// DeleteDocument deletes a document by its ID
func (c *Client) DeleteDocument(ctx context.Context, index, id string) error {
	req := opensearchapi.DeleteRequest{
		Index:      index,
		DocumentID: id,
		Refresh:    "true",
	}

	res, err := req.Do(ctx, c.client)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == 404 {
			return fmt.Errorf("document not found")
		}
		return fmt.Errorf("delete request failed with status: %s", res.Status())
	}

	return nil
}

// SearchDocuments performs a search query on an index
func (c *Client) SearchDocuments(ctx context.Context, index string, query map[string]interface{}) ([]map[string]interface{}, error) {
	body, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	req := opensearchapi.SearchRequest{
		Index: []string{index},
		Body:  bytes.NewReader(body),
	}

	res, err := req.Do(ctx, c.client)
	if err != nil {
		return nil, fmt.Errorf("failed to search documents: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("search request failed with status: %s", res.Status())
	}

	var response SearchResponse
	if err := parseResponse(res.Body, &response); err != nil {
		return nil, err
	}

	results := make([]map[string]interface{}, 0, len(response.Hits.Hits))
	for _, hit := range response.Hits.Hits {
		doc := hit.Source
		doc["_id"] = hit.ID
		doc["_score"] = hit.Score
		results = append(results, doc)
	}

	return results, nil
}

// SearchAll retrieves all documents from an index using match_all query
func (c *Client) SearchAll(ctx context.Context, index string) ([]map[string]interface{}, error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
	}
	return c.SearchDocuments(ctx, index, query)
}

// CreateIndex creates a new index with optional settings and mappings
func (c *Client) CreateIndex(ctx context.Context, index string, body map[string]interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal index body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req := opensearchapi.IndicesCreateRequest{
		Index: index,
		Body:  bodyReader,
	}

	res, err := req.Do(ctx, c.client)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("create index request failed with status: %s", res.Status())
	}

	return nil
}

// DeleteIndex deletes an index
func (c *Client) DeleteIndex(ctx context.Context, index string) error {
	req := opensearchapi.IndicesDeleteRequest{
		Index: []string{index},
	}

	res, err := req.Do(ctx, c.client)
	if err != nil {
		return fmt.Errorf("failed to delete index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == 404 {
			return fmt.Errorf("index not found")
		}
		return fmt.Errorf("delete index request failed with status: %s", res.Status())
	}

	return nil
}

// IndexExists checks if an index exists
func (c *Client) IndexExists(ctx context.Context, index string) (bool, error) {
	req := opensearchapi.IndicesExistsRequest{
		Index: []string{index},
	}

	res, err := req.Do(ctx, c.client)
	if err != nil {
		return false, fmt.Errorf("failed to check index existence: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode == 404 {
		return false, nil
	}

	if res.IsError() {
		return false, fmt.Errorf("index exists request failed with status: %s", res.Status())
	}

	return true, nil
}

// BulkCreate performs bulk indexing of multiple documents
func (c *Client) BulkCreate(ctx context.Context, index string, documents []map[string]interface{}) error {
	if len(documents) == 0 {
		return nil
	}

	var buf bytes.Buffer
	for _, doc := range documents {
		// Action line
		action := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": index,
			},
		}
		if id, ok := doc["_id"]; ok {
			action["index"].(map[string]interface{})["_id"] = id
			delete(doc, "_id")
		}

		actionBytes, err := json.Marshal(action)
		if err != nil {
			return fmt.Errorf("failed to marshal bulk action: %w", err)
		}
		buf.Write(actionBytes)
		buf.WriteByte('\n')

		// Document line
		docBytes, err := json.Marshal(doc)
		if err != nil {
			return fmt.Errorf("failed to marshal document: %w", err)
		}
		buf.Write(docBytes)
		buf.WriteByte('\n')
	}

	req := opensearchapi.BulkRequest{
		Body:    &buf,
		Refresh: "true",
	}

	res, err := req.Do(ctx, c.client)
	if err != nil {
		return fmt.Errorf("failed to perform bulk operation: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("bulk request failed with status: %s", res.Status())
	}

	var response BulkResponse
	if err := parseResponse(res.Body, &response); err != nil {
		return err
	}

	if response.Errors {
		var errorMessages []string
		for _, item := range response.Items {
			for _, op := range item {
				if op.Error.Type != "" {
					errorMessages = append(errorMessages, fmt.Sprintf("%s: %s", op.Error.Type, op.Error.Reason))
				}
			}
		}
		if len(errorMessages) > 0 {
			return fmt.Errorf("bulk operation had errors: %s", strings.Join(errorMessages, "; "))
		}
	}

	return nil
}