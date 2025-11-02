package opensearch

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

// TestParseResponse tests the parseResponse helper function
func TestParseResponse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		target  interface{}
		want    interface{}
		wantErr bool
	}{
		{
			name:  "valid JSON - GetResponse",
			input: `{"_index":"test","_id":"1","_version":1,"found":true,"_source":{"name":"test"}}`,
			target: &GetResponse{},
			want: &GetResponse{
				Index:   "test",
				ID:      "1",
				Version: 1,
				Found:   true,
				Source:  map[string]interface{}{"name": "test"},
			},
			wantErr: false,
		},
		{
			name:  "valid JSON - IndexResponse",
			input: `{"_index":"test","_id":"1","_version":1,"result":"created"}`,
			target: &IndexResponse{},
			want: &IndexResponse{
				Index:   "test",
				ID:      "1",
				Version: 1,
				Result:  "created",
			},
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			input:   `{"invalid json`,
			target:  &GetResponse{},
			want:    &GetResponse{},
			wantErr: true,
		},
		{
			name:  "empty JSON object",
			input: `{}`,
			target: &GetResponse{},
			want: &GetResponse{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			err := parseResponse(reader, tt.target)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !reflect.DeepEqual(tt.target, tt.want) {
				t.Errorf("parseResponse() got = %v, want %v", tt.target, tt.want)
			}
		})
	}
}

// TestMatchAllQuery tests the MatchAllQuery builder
func TestMatchAllQuery(t *testing.T) {
	result := MatchAllQuery()

	expected := map[string]interface{}{
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("MatchAllQuery() = %v, want %v", result, expected)
	}

	// Verify JSON structure
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to marshal result: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}
}

// TestMatchQuery tests the MatchQuery builder
func TestMatchQuery(t *testing.T) {
	tests := []struct {
		name  string
		field string
		value string
		want  map[string]interface{}
	}{
		{
			name:  "simple match query",
			field: "title",
			value: "test",
			want: map[string]interface{}{
				"query": map[string]interface{}{
					"match": map[string]interface{}{
						"title": "test",
					},
				},
			},
		},
		{
			name:  "match query with dots in field",
			field: "user.name",
			value: "john",
			want: map[string]interface{}{
				"query": map[string]interface{}{
					"match": map[string]interface{}{
						"user.name": "john",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MatchQuery(tt.field, tt.value)
			if !reflect.DeepEqual(result, tt.want) {
				t.Errorf("MatchQuery() = %v, want %v", result, tt.want)
			}
		})
	}
}

// TestNotMatchQuery tests the NotMatchQuery builder
func TestNotMatchQuery(t *testing.T) {
	result := NotMatchQuery("status", "inactive")

	// Check structure
	query, ok := result["query"].(map[string]interface{})
	if !ok {
		t.Fatal("query is not a map")
	}

	boolQuery, ok := query["bool"].(map[string]interface{})
	if !ok {
		t.Fatal("bool is not a map")
	}

	mustNot, ok := boolQuery["must_not"].([]map[string]interface{})
	if !ok {
		t.Fatal("must_not is not a slice")
	}

	if len(mustNot) != 1 {
		t.Errorf("must_not length = %d, want 1", len(mustNot))
	}

	matchClause, ok := mustNot[0]["match"].(map[string]interface{})
	if !ok {
		t.Fatal("match is not a map")
	}

	if matchClause["status"] != "inactive" {
		t.Errorf("match value = %v, want 'inactive'", matchClause["status"])
	}
}

// TestMatchMapQuery tests the MatchMapQuery builder
func TestMatchMapQuery(t *testing.T) {
	tests := []struct {
		name        string
		fieldValues map[string]interface{}
		wantCount   int
	}{
		{
			name: "single field",
			fieldValues: map[string]interface{}{
				"title": "test",
			},
			wantCount: 1,
		},
		{
			name: "multiple fields",
			fieldValues: map[string]interface{}{
				"title":  "test",
				"status": "active",
				"count":  5,
			},
			wantCount: 3,
		},
		{
			name:        "empty map",
			fieldValues: map[string]interface{}{},
			wantCount:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MatchMapQuery(tt.fieldValues)

			query, ok := result["query"].(map[string]interface{})
			if !ok {
				t.Fatal("query is not a map")
			}

			boolQuery, ok := query["bool"].(map[string]interface{})
			if !ok {
				t.Fatal("bool is not a map")
			}

			must, ok := boolQuery["must"].([]map[string]interface{})
			if !ok {
				t.Fatal("must is not a slice")
			}

			if len(must) != tt.wantCount {
				t.Errorf("must clause count = %d, want %d", len(must), tt.wantCount)
			}

			// Verify all fields are present
			for _, clause := range must {
				match, ok := clause["match"].(map[string]interface{})
				if !ok {
					t.Fatal("match is not a map")
				}
				if len(match) != 1 {
					t.Errorf("match should have exactly 1 field, got %d", len(match))
				}
			}
		})
	}
}

// TestNotMatchMapQuery tests the NotMatchMapQuery builder
func TestNotMatchMapQuery(t *testing.T) {
	fieldValues := map[string]interface{}{
		"status": "deleted",
		"hidden": true,
	}

	result := NotMatchMapQuery(fieldValues)

	query, ok := result["query"].(map[string]interface{})
	if !ok {
		t.Fatal("query is not a map")
	}

	boolQuery, ok := query["bool"].(map[string]interface{})
	if !ok {
		t.Fatal("bool is not a map")
	}

	mustNot, ok := boolQuery["must_not"].([]map[string]interface{})
	if !ok {
		t.Fatal("must_not is not a slice")
	}

	if len(mustNot) != 2 {
		t.Errorf("must_not clause count = %d, want 2", len(mustNot))
	}
}

// TestTermQuery tests the TermQuery builder
func TestTermQuery(t *testing.T) {
	tests := []struct {
		name  string
		field string
		value interface{}
	}{
		{
			name:  "string value",
			field: "status",
			value: "active",
		},
		{
			name:  "integer value",
			field: "count",
			value: 42,
		},
		{
			name:  "boolean value",
			field: "enabled",
			value: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TermQuery(tt.field, tt.value)

			query, ok := result["query"].(map[string]interface{})
			if !ok {
				t.Fatal("query is not a map")
			}

			term, ok := query["term"].(map[string]interface{})
			if !ok {
				t.Fatal("term is not a map")
			}

			if term[tt.field] != tt.value {
				t.Errorf("term value = %v, want %v", term[tt.field], tt.value)
			}
		})
	}
}

// TestNotTermQuery tests the NotTermQuery builder
func TestNotTermQuery(t *testing.T) {
	result := NotTermQuery("status.keyword", "deleted")

	query, ok := result["query"].(map[string]interface{})
	if !ok {
		t.Fatal("query is not a map")
	}

	boolQuery, ok := query["bool"].(map[string]interface{})
	if !ok {
		t.Fatal("bool is not a map")
	}

	mustNot, ok := boolQuery["must_not"].([]map[string]interface{})
	if !ok {
		t.Fatal("must_not is not a slice")
	}

	if len(mustNot) != 1 {
		t.Errorf("must_not length = %d, want 1", len(mustNot))
	}
}

// TestRangeQuery tests the RangeQuery builder
func TestRangeQuery(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		gte      interface{}
		lte      interface{}
		wantGte  bool
		wantLte  bool
	}{
		{
			name:    "both gte and lte",
			field:   "age",
			gte:     18,
			lte:     65,
			wantGte: true,
			wantLte: true,
		},
		{
			name:    "only gte",
			field:   "price",
			gte:     100,
			lte:     nil,
			wantGte: true,
			wantLte: false,
		},
		{
			name:    "only lte",
			field:   "score",
			gte:     nil,
			lte:     100,
			wantGte: false,
			wantLte: true,
		},
		{
			name:    "neither gte nor lte",
			field:   "value",
			gte:     nil,
			lte:     nil,
			wantGte: false,
			wantLte: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RangeQuery(tt.field, tt.gte, tt.lte)

			query, ok := result["query"].(map[string]interface{})
			if !ok {
				t.Fatal("query is not a map")
			}

			rangeQuery, ok := query["range"].(map[string]interface{})
			if !ok {
				t.Fatal("range is not a map")
			}

			rangeCondition, ok := rangeQuery[tt.field].(map[string]interface{})
			if !ok {
				t.Fatal("range condition is not a map")
			}

			if tt.wantGte {
				if rangeCondition["gte"] != tt.gte {
					t.Errorf("gte = %v, want %v", rangeCondition["gte"], tt.gte)
				}
			} else {
				if _, exists := rangeCondition["gte"]; exists {
					t.Error("gte should not exist")
				}
			}

			if tt.wantLte {
				if rangeCondition["lte"] != tt.lte {
					t.Errorf("lte = %v, want %v", rangeCondition["lte"], tt.lte)
				}
			} else {
				if _, exists := rangeCondition["lte"]; exists {
					t.Error("lte should not exist")
				}
			}
		})
	}
}

// TestBoolQuery tests the BoolQuery builder
func TestBoolQuery(t *testing.T) {
	tests := []struct {
		name    string
		must    []map[string]interface{}
		should  []map[string]interface{}
		mustNot []map[string]interface{}
		wantMust    bool
		wantShould  bool
		wantMustNot bool
	}{
		{
			name: "all clauses present",
			must: []map[string]interface{}{
				{"match": map[string]interface{}{"title": "test"}},
			},
			should: []map[string]interface{}{
				{"term": map[string]interface{}{"status": "active"}},
			},
			mustNot: []map[string]interface{}{
				{"term": map[string]interface{}{"deleted": true}},
			},
			wantMust:    true,
			wantShould:  true,
			wantMustNot: true,
		},
		{
			name: "only must clause",
			must: []map[string]interface{}{
				{"match": map[string]interface{}{"title": "test"}},
			},
			should:      nil,
			mustNot:     nil,
			wantMust:    true,
			wantShould:  false,
			wantMustNot: false,
		},
		{
			name:        "empty clauses",
			must:        []map[string]interface{}{},
			should:      []map[string]interface{}{},
			mustNot:     []map[string]interface{}{},
			wantMust:    false,
			wantShould:  false,
			wantMustNot: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BoolQuery(tt.must, tt.should, tt.mustNot)

			query, ok := result["query"].(map[string]interface{})
			if !ok {
				t.Fatal("query is not a map")
			}

			boolQuery, ok := query["bool"].(map[string]interface{})
			if !ok {
				t.Fatal("bool is not a map")
			}

			if tt.wantMust {
				if _, exists := boolQuery["must"]; !exists {
					t.Error("must clause should exist")
				}
			} else {
				if _, exists := boolQuery["must"]; exists {
					t.Error("must clause should not exist")
				}
			}

			if tt.wantShould {
				if _, exists := boolQuery["should"]; !exists {
					t.Error("should clause should exist")
				}
			} else {
				if _, exists := boolQuery["should"]; exists {
					t.Error("should clause should not exist")
				}
			}

			if tt.wantMustNot {
				if _, exists := boolQuery["must_not"]; !exists {
					t.Error("must_not clause should exist")
				}
			} else {
				if _, exists := boolQuery["must_not"]; exists {
					t.Error("must_not clause should not exist")
				}
			}
		})
	}
}

// TestWithSize tests the WithSize modifier
func TestWithSize(t *testing.T) {
	query := MatchAllQuery()
	result := WithSize(query, 50)

	if result["size"] != 50 {
		t.Errorf("size = %v, want 50", result["size"])
	}

	// Verify query is still intact
	if _, exists := result["query"]; !exists {
		t.Error("query should still exist after adding size")
	}
}

// TestWithFrom tests the WithFrom modifier
func TestWithFrom(t *testing.T) {
	query := MatchAllQuery()
	result := WithFrom(query, 100)

	if result["from"] != 100 {
		t.Errorf("from = %v, want 100", result["from"])
	}

	// Verify query is still intact
	if _, exists := result["query"]; !exists {
		t.Error("query should still exist after adding from")
	}
}

// TestWithSort tests the WithSort modifier
func TestWithSort(t *testing.T) {
	tests := []struct {
		name  string
		field string
		order string
	}{
		{
			name:  "ascending sort",
			field: "created_at",
			order: "asc",
		},
		{
			name:  "descending sort",
			field: "score",
			order: "desc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := MatchAllQuery()
			result := WithSort(query, tt.field, tt.order)

			sort, ok := result["sort"].([]map[string]interface{})
			if !ok {
				t.Fatal("sort is not a slice")
			}

			if len(sort) != 1 {
				t.Errorf("sort length = %d, want 1", len(sort))
			}

			sortField, ok := sort[0][tt.field].(map[string]interface{})
			if !ok {
				t.Fatal("sort field is not a map")
			}

			if sortField["order"] != tt.order {
				t.Errorf("sort order = %v, want %v", sortField["order"], tt.order)
			}
		})
	}
}

// TestQueryChaining tests chaining multiple modifiers
func TestQueryChaining(t *testing.T) {
	query := MatchQuery("title", "golang")
	query = WithSize(query, 20)
	query = WithFrom(query, 10)
	query = WithSort(query, "created_at", "desc")

	// Verify all parameters exist
	if query["size"] != 20 {
		t.Errorf("size = %v, want 20", query["size"])
	}

	if query["from"] != 10 {
		t.Errorf("from = %v, want 10", query["from"])
	}

	sort, ok := query["sort"].([]map[string]interface{})
	if !ok {
		t.Fatal("sort is not a slice")
	}

	if len(sort) != 1 {
		t.Errorf("sort length = %d, want 1", len(sort))
	}

	// Verify original query is intact
	if _, exists := query["query"]; !exists {
		t.Error("query should exist after chaining modifiers")
	}
}

// TestJSONMarshaling tests that all query builders produce valid JSON
func TestJSONMarshaling(t *testing.T) {
	queries := []struct {
		name  string
		query map[string]interface{}
	}{
		{"MatchAllQuery", MatchAllQuery()},
		{"MatchQuery", MatchQuery("field", "value")},
		{"NotMatchQuery", NotMatchQuery("field", "value")},
		{"TermQuery", TermQuery("field", "value")},
		{"NotTermQuery", NotTermQuery("field", "value")},
		{"RangeQuery", RangeQuery("field", 1, 10)},
		{"BoolQuery", BoolQuery(
			[]map[string]interface{}{{"match": map[string]interface{}{"f": "v"}}},
			nil,
			nil,
		)},
		{"MatchMapQuery", MatchMapQuery(map[string]interface{}{"f1": "v1", "f2": "v2"})},
		{"NotMatchMapQuery", NotMatchMapQuery(map[string]interface{}{"f1": "v1"})},
	}

	for _, tt := range queries {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal to JSON
			jsonBytes, err := json.Marshal(tt.query)
			if err != nil {
				t.Fatalf("Failed to marshal %s: %v", tt.name, err)
			}

			// Unmarshal back
			var decoded map[string]interface{}
			if err := json.Unmarshal(jsonBytes, &decoded); err != nil {
				t.Fatalf("Failed to unmarshal %s: %v", tt.name, err)
			}

			// Verify it has a query key (except for modifiers)
			if _, exists := decoded["query"]; !exists {
				t.Errorf("%s should have a 'query' key", tt.name)
			}
		})
	}
}

// TestParseResponseWithDifferentTypes tests parseResponse with various response types
func TestParseResponseWithDifferentTypes(t *testing.T) {
	t.Run("SearchResponse", func(t *testing.T) {
		input := `{
			"took": 5,
			"hits": {
				"total": {"value": 1, "relation": "eq"},
				"max_score": 1.0,
				"hits": [
					{
						"_index": "test",
						"_id": "1",
						"_score": 1.0,
						"_source": {"title": "test"}
					}
				]
			}
		}`

		var response SearchResponse
		err := parseResponse(bytes.NewReader([]byte(input)), &response)
		if err != nil {
			t.Fatalf("parseResponse failed: %v", err)
		}

		if response.Took != 5 {
			t.Errorf("took = %d, want 5", response.Took)
		}

		if response.Hits.Total.Value != 1 {
			t.Errorf("total.value = %d, want 1", response.Hits.Total.Value)
		}

		if len(response.Hits.Hits) != 1 {
			t.Errorf("hits length = %d, want 1", len(response.Hits.Hits))
		}
	})

	t.Run("ErrorResponse", func(t *testing.T) {
		input := `{
			"error": {
				"type": "index_not_found_exception",
				"reason": "no such index [missing]"
			},
			"status": 404
		}`

		var response ErrorResponse
		err := parseResponse(bytes.NewReader([]byte(input)), &response)
		if err != nil {
			t.Fatalf("parseResponse failed: %v", err)
		}

		if response.Status != 404 {
			t.Errorf("status = %d, want 404", response.Status)
		}

		if response.Error.Type != "index_not_found_exception" {
			t.Errorf("error type = %s, want 'index_not_found_exception'", response.Error.Type)
		}
	})
}