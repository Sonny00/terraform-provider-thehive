package client

import (
	"context"
	"fmt"
	"net/http"
)

type CustomField struct {
	ID          string   `json:"_id"`
	Name        string   `json:"name"`
	DisplayName string   `json:"displayName"`
	Description string   `json:"description"`
	Group       string   `json:"group"`
	Type        string   `json:"type"`
	Options     []string `json:"options,omitempty"`
	Mandatory   bool     `json:"mandatory"`
}

type CustomFieldInput struct {
	Name        string   `json:"name"`
	DisplayName string   `json:"displayName,omitempty"`
	Description string   `json:"description,omitempty"`
	Group       string   `json:"group"`
	Type        string   `json:"type"`
	Options     []string `json:"options,omitempty"`
	Mandatory   bool     `json:"mandatory,omitempty"`
}

type CustomFieldUpdateInput struct {
	DisplayName string   `json:"displayName,omitempty"`
	Description string   `json:"description,omitempty"`
	Options     []string `json:"options,omitempty"`
	Mandatory   bool     `json:"mandatory,omitempty"`
}

const customFieldBasePath = "/api/v1/customField"

func (c *Client) CreateCustomField(ctx context.Context, input CustomFieldInput) (*CustomField, error) {
	var cf CustomField
	if err := c.doJSON(ctx, http.MethodPost, customFieldBasePath, input, &cf, http.StatusCreated); err != nil {
		return nil, fmt.Errorf("creating custom field: %w", err)
	}
	return &cf, nil
}

func (c *Client) GetCustomField(ctx context.Context, id string) (*CustomField, error) {
	body := map[string]any{
		"query": []map[string]any{
			{"_name": "getCustomField", "idOrName": id},
		},
	}
	var results []CustomField
	if err := c.doJSON(ctx, http.MethodPost, "/api/v1/query", body, &results, http.StatusOK); err != nil {
		return nil, fmt.Errorf("reading custom field %q: %w", id, err)
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("custom field %q not found", id)
	}
	return &results[0], nil
}

func (c *Client) UpdateCustomField(ctx context.Context, id string, input CustomFieldUpdateInput) (*CustomField, error) {
	var cf CustomField
	if err := c.doJSON(ctx, http.MethodPatch, fmt.Sprintf("%s/%s", customFieldBasePath, id), input, &cf, http.StatusOK); err != nil {
		return nil, fmt.Errorf("updating custom field %q: %w", id, err)
	}
	return &cf, nil
}

func (c *Client) DeleteCustomField(ctx context.Context, id string) error {
	if err := c.doJSON(ctx, http.MethodDelete, fmt.Sprintf("%s/%s", customFieldBasePath, id), nil, nil, http.StatusNoContent); err != nil {
		return fmt.Errorf("deleting custom field %q: %w", id, err)
	}
	return nil
}
