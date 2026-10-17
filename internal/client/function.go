package client

import (
	"context"
	"fmt"
	"net/http"
)

type FunctionInput struct {
	Name       string            `json:"name"`
	Mode       string            `json:"mode"`
	Definition string            `json:"definition"`
	Config     map[string]string `json:"config,omitempty"`
	Types      []string          `json:"types"`
}

type FunctionUpdateInput struct {
	Name       string            `json:"name,omitempty"`
	Mode       string            `json:"mode,omitempty"`
	Definition string            `json:"definition,omitempty"`
	Config     map[string]string `json:"config"`
	Types      []string          `json:"types"`
}

type Function struct {
	ID         string            `json:"_id"`
	Type       string            `json:"_type"`
	Name       string            `json:"name"`
	Mode       string            `json:"mode"`
	Definition string            `json:"definition"`
	Config     map[string]string `json:"config"`
	Types      []string          `json:"types"`
	CreatedAt  int64             `json:"_createdAt,omitempty"`
	CreatedBy  string            `json:"_createdBy,omitempty"`
	UpdatedAt  *int64            `json:"_updatedAt,omitempty"`
	UpdatedBy  string            `json:"_updatedBy,omitempty"`
}

const functionBasePath = "/api/v1/function"

func (c *Client) CreateFunction(ctx context.Context, input FunctionInput) (*Function, error) {
	var function Function
	if err := c.doJSON(ctx, http.MethodPost, functionBasePath, input, &function, http.StatusCreated); err != nil {
		return nil, fmt.Errorf("creating function: %w", err)
	}
	return &function, nil
}

func (c *Client) GetFunction(ctx context.Context, id string) (*Function, error) {
	body := map[string]any{
		"query": []map[string]any{
			{"_name": "getFunction", "idOrName": id},
		},
	}
	var results []Function
	if err := c.doJSON(ctx, http.MethodPost, "/api/v1/query?name=get-function", body, &results, http.StatusOK); err != nil {
		return nil, fmt.Errorf("reading function %q: %w", id, err)
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("function %q not found", id)
	}
	return &results[0], nil
}

func (c *Client) UpdateFunction(ctx context.Context, id string, input FunctionUpdateInput) (*Function, error) {
	body, status, err := c.doRequest(ctx, http.MethodPatch, fmt.Sprintf("%s/%s", functionBasePath, id), input)
	if err != nil {
		return nil, fmt.Errorf("updating function %q: %w", id, err)
	}
	if status != http.StatusOK && status != http.StatusNoContent {
		return nil, fmt.Errorf("updating function %q: %w", id, parseAPIError(status, body))
	}
	return c.GetFunction(ctx, id)
}

func (c *Client) DeleteFunction(ctx context.Context, id string) error {
	if err := c.doJSON(ctx, http.MethodDelete, fmt.Sprintf("%s/%s", functionBasePath, id), nil, nil, http.StatusNoContent); err != nil {
		return fmt.Errorf("deleting function %q: %w", id, err)
	}
	return nil
}
