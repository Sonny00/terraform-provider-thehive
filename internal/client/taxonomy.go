package client

import (
	"context"
	"fmt"
	"net/http"
)

type TaxonomyPredicate struct {
	Value       string `json:"value"`
	Expanded    string `json:"expanded,omitempty"`
	Exclusive   bool   `json:"exclusive,omitempty"`
	Description string `json:"description,omitempty"`
	Colour      string `json:"colour,omitempty"`
}

type TaxonomyEntry struct {
	Value          string `json:"value"`
	Expanded       string `json:"expanded,omitempty"`
	Colour         string `json:"colour,omitempty"`
	Description    string `json:"description,omitempty"`
	NumericalValue *int64 `json:"numerical_value,omitempty"`
}

type TaxonomyValue struct {
	Predicate string          `json:"predicate"`
	Entry     []TaxonomyEntry `json:"entry"`
}

type TaxonomyInput struct {
	Namespace   string              `json:"namespace"`
	Description string              `json:"description,omitempty"`
	Version     int64               `json:"version,omitempty"`
	Exclusive   bool                `json:"exclusive,omitempty"`
	Predicates  []TaxonomyPredicate `json:"predicates"`
	Values      []TaxonomyValue     `json:"values,omitempty"`
}

type Taxonomy struct {
	ID          string `json:"_id"`
	Namespace   string `json:"namespace"`
	Description string `json:"description"`
	Version     int64  `json:"version"`
	CreatedAt   int64  `json:"_createdAt,omitempty"`
	CreatedBy   string `json:"_createdBy,omitempty"`
}

const taxonomyBasePath = "/api/v1/taxonomy"

func (c *Client) CreateTaxonomy(ctx context.Context, input TaxonomyInput) (*Taxonomy, error) {
	var taxonomy Taxonomy
	if err := c.doJSON(ctx, http.MethodPost, taxonomyBasePath, input, &taxonomy, http.StatusCreated); err != nil {
		return nil, fmt.Errorf("creating taxonomy: %w", err)
	}
	return &taxonomy, nil
}

func (c *Client) GetTaxonomy(ctx context.Context, id string) (*Taxonomy, error) {
	body := map[string]any{
		"query": []map[string]any{
			{"_name": "getTaxonomy", "idOrName": id},
		},
	}
	var results []Taxonomy
	if err := c.doJSON(ctx, http.MethodPost, "/api/v1/query?name=get-taxonomy", body, &results, http.StatusOK); err != nil {
		return nil, fmt.Errorf("reading taxonomy %q: %w", id, err)
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("taxonomy %q not found", id)
	}
	return &results[0], nil
}

func (c *Client) DeleteTaxonomy(ctx context.Context, id string) error {
	if err := c.doJSON(ctx, http.MethodDelete, fmt.Sprintf("%s/%s", taxonomyBasePath, id), nil, nil, http.StatusNoContent); err != nil {
		return fmt.Errorf("deleting taxonomy %q: %w", id, err)
	}
	return nil
}
