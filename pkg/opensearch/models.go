package opensearch

import (
	"encoding/json"
	"fmt"
	"io"
)

// GetResponse represents the response from a GET document request
type GetResponse struct {
	Index   string                 `json:"_index"`
	ID      string                 `json:"_id"`
	Version int                    `json:"_version"`
	Found   bool                   `json:"found"`
	Source  map[string]interface{} `json:"_source"`
}

// SearchResponse represents the response from a search request
type SearchResponse struct {
	Took int `json:"took"`
	Hits struct {
		Total struct {
			Value    int    `json:"value"`
			Relation string `json:"relation"`
		} `json:"total"`
		MaxScore float64 `json:"max_score"`
		Hits     []Hit   `json:"hits"`
	} `json:"hits"`
}

// Hit represents a single search result
type Hit struct {
	Index  string                 `json:"_index"`
	ID     string                 `json:"_id"`
	Score  float64                `json:"_score"`
	Source map[string]interface{} `json:"_source"`
}

// BulkResponse represents the response from a bulk request
type BulkResponse struct {
	Took   int                   `json:"took"`
	Errors bool                  `json:"errors"`
	Items  []map[string]BulkItem `json:"items"`
}

// BulkItem represents a single item in a bulk response
type BulkItem struct {
	Index   string `json:"_index"`
	ID      string `json:"_id"`
	Version int    `json:"_version"`
	Result  string `json:"result"`
	Status  int    `json:"status"`
	Error   struct {
		Type   string `json:"type"`
		Reason string `json:"reason"`
	} `json:"error"`
}

// IndexResponse represents the response from an index operation
type IndexResponse struct {
	Index   string `json:"_index"`
	ID      string `json:"_id"`
	Version int    `json:"_version"`
	Result  string `json:"result"`
}

// DeleteResponse represents the response from a delete operation
type DeleteResponse struct {
	Index   string `json:"_index"`
	ID      string `json:"_id"`
	Version int    `json:"_version"`
	Result  string `json:"result"`
}

// UpdateResponse represents the response from an update operation
type UpdateResponse struct {
	Index   string `json:"_index"`
	ID      string `json:"_id"`
	Version int    `json:"_version"`
	Result  string `json:"result"`
}

// ErrorResponse represents an error response from OpenSearch
type ErrorResponse struct {
	Error struct {
		Type   string `json:"type"`
		Reason string `json:"reason"`
	} `json:"error"`
	Status int `json:"status"`
}

// parseResponse is a helper function to parse JSON responses
func parseResponse(body io.Reader, v interface{}) error {
	if err := json.NewDecoder(body).Decode(v); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}
	return nil
}

// Query builders for common search patterns

// MatchAllQuery creates a match_all query
func MatchAllQuery() map[string]interface{} {
	return map[string]interface{}{
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
	}
}

// MatchQuery creates a match query for a specific field
func MatchQuery(field, value string) map[string]interface{} {
	return map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				field: value,
			},
		},
	}
}

// NotMatchQuery creates a bool query that excludes documents matching the specified field and value
func NotMatchQuery(field, value string) map[string]interface{} {
	return map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must_not": []map[string]interface{}{
					{
						"match": map[string]interface{}{
							field: value,
						},
					},
				},
			},
		},
	}
}

// MatchMapQuery creates a bool query with must clause matching all field-value pairs in the map
func MatchMapQuery(fieldValues map[string]interface{}) map[string]interface{} {
	mustClauses := make([]map[string]interface{}, 0, len(fieldValues))

	for field, value := range fieldValues {
		mustClauses = append(mustClauses, map[string]interface{}{
			"match": map[string]interface{}{
				field: value,
			},
		})
	}

	return map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": mustClauses,
			},
		},
	}
}

// TermQuery creates a term query for exact matching
func TermQuery(field string, value interface{}) map[string]interface{} {
	return map[string]interface{}{
		"query": map[string]interface{}{
			"term": map[string]interface{}{
				field: value,
			},
		},
	}
}

// NotTermQuery creates a bool query that excludes documents with exact field value match
func NotTermQuery(field string, value interface{}) map[string]interface{} {
	return map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must_not": []map[string]interface{}{
					{
						"term": map[string]interface{}{
							field: value,
						},
					},
				},
			},
		},
	}
}

// RangeQuery creates a range query
func RangeQuery(field string, gte, lte interface{}) map[string]interface{} {
	rangeCondition := make(map[string]interface{})
	if gte != nil {
		rangeCondition["gte"] = gte
	}
	if lte != nil {
		rangeCondition["lte"] = lte
	}

	return map[string]interface{}{
		"query": map[string]interface{}{
			"range": map[string]interface{}{
				field: rangeCondition,
			},
		},
	}
}

// BoolQuery creates a bool query for complex queries
func BoolQuery(must, should, mustNot []map[string]interface{}) map[string]interface{} {
	boolQuery := make(map[string]interface{})

	if len(must) > 0 {
		boolQuery["must"] = must
	}
	if len(should) > 0 {
		boolQuery["should"] = should
	}
	if len(mustNot) > 0 {
		boolQuery["must_not"] = mustNot
	}

	return map[string]interface{}{
		"query": map[string]interface{}{
			"bool": boolQuery,
		},
	}
}

// WithSize adds a size parameter to a query
func WithSize(query map[string]interface{}, size int) map[string]interface{} {
	query["size"] = size
	return query
}

// WithFrom adds a from parameter to a query (for pagination)
func WithFrom(query map[string]interface{}, from int) map[string]interface{} {
	query["from"] = from
	return query
}

// WithSort adds sorting to a query
func WithSort(query map[string]interface{}, field, order string) map[string]interface{} {
	query["sort"] = []map[string]interface{}{
		{
			field: map[string]interface{}{
				"order": order,
			},
		},
	}
	return query
}

