package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
)

type SearchRepository struct {
	es *elasticsearch.Client
}

func NewSearchRepository(es *elasticsearch.Client) *SearchRepository {
	return &SearchRepository{es: es}
}

func (s *SearchRepository) FuzzyMatch(ctx context.Context, tenantID, firstName, lastName string, dob time.Time) ([]string, error) {
	// Build the Elasticsearch boolean query
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{
						"term": map[string]interface{}{
							"tenant_id": tenantID,
						},
					},
					{
						"range": map[string]interface{}{
							"dob": map[string]interface{}{
								"gte": dob.Add(-365 * 24 * time.Hour), // +/- 1 year fuzziness
								"lte": dob.Add(365 * 24 * time.Hour),
							},
						},
					},
				},
				"should": []map[string]interface{}{
					{
						"match": map[string]interface{}{
							"first_name": map[string]interface{}{
								"query":     firstName,
								"fuzziness": "AUTO",
							},
						},
					},
					{
						"match": map[string]interface{}{
							"last_name": map[string]interface{}{
								"query":     lastName,
								"fuzziness": "AUTO",
							},
						},
					},
				},
				"minimum_should_match": 1,
			},
		},
		"_source": []string{"party_id"},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, fmt.Errorf("failed to encode search query: %w", err)
	}

	// Perform the search
	res, err := s.es.Search(
		s.es.Search.WithContext(ctx),
		s.es.Search.WithIndex("parties"),
		s.es.Search.WithBody(&buf),
		s.es.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		slog.Error("Elasticsearch query failed", "error", err)
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("elasticsearch error response: %s", res.String())
	}

	// Parse the response
	var response struct {
		Hits struct {
			Hits []struct {
				ID string `json:"_id"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	var candidates []string
	for _, hit := range response.Hits.Hits {
		candidates = append(candidates, hit.ID)
	}

	return candidates, nil
}
