package client

import (
	"context"
	"fmt"
	"net/http"
)

type ObservableType struct {
	ID           string `json:"_id"`
	Name         string `json:"name"`
	IsAttachment bool   `json:"isAttachment"`
}

type ObservableTypeInput struct {
	Name         string `json:"name"`
	IsAttachment bool   `json:"isAttachment,omitempty"`
}

const observableTypeBasePath = "/api/v1/observable/type"

func (c *Client) CreateObservableType(ctx context.Context, input ObservableTypeInput) (*ObservableType, error) {
	var ot ObservableType
	if err := c.doJSON(ctx, http.MethodPost, observableTypeBasePath, input, &ot, http.StatusCreated); err != nil {
		return nil, fmt.Errorf("creating observable type: %w", err)
	}
	return &ot, nil
}

func (c *Client) GetObservableType(ctx context.Context, id string) (*ObservableType, error) {
	body := map[string]any{
		"query": []map[string]any{
			{"_name": "getObservableType", "idOrName": id},
		},
	}
	var results []ObservableType
	if err := c.doJSON(ctx, http.MethodPost, "/api/v1/query", body, &results, http.StatusOK); err != nil {
		return nil, fmt.Errorf("reading observable type %q: %w", id, err)
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("observable type %q not found", id)
	}
	return &results[0], nil
}

func (c *Client) DeleteObservableType(ctx context.Context, id string) error {
	if err := c.doJSON(ctx, http.MethodDelete, fmt.Sprintf("%s/%s", observableTypeBasePath, id), nil, nil, http.StatusNoContent); err != nil {
		return fmt.Errorf("deleting observable type %q: %w", id, err)
	}
	return nil
}
