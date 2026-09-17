package client

import (
	"context"
	"fmt"
	"net/http"
)

type PageTemplate struct {
	ID       string `json:"_id"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	Category string `json:"category"`
	Order    int    `json:"order"`
}

type PageTemplateInput struct {
	Title    string `json:"title"`
	Content  string `json:"content,omitempty"`
	Category string `json:"category,omitempty"`
	Order    int    `json:"order,omitempty"`
}

const pageTemplateBasePath = "/api/v1/pageTemplate"

func (c *Client) CreatePageTemplate(ctx context.Context, input PageTemplateInput) (*PageTemplate, error) {
	var pt PageTemplate
	if err := c.doJSON(ctx, http.MethodPost, pageTemplateBasePath, input, &pt, http.StatusCreated); err != nil {
		return nil, fmt.Errorf("creating page template: %w", err)
	}
	return &pt, nil
}

func (c *Client) GetPageTemplate(ctx context.Context, id string) (*PageTemplate, error) {
	body := map[string]any{
		"query": []map[string]any{
			{"_name": "getPageTemplate", "idOrName": id},
		},
	}
	var results []PageTemplate
	if err := c.doJSON(ctx, http.MethodPost, "/api/v1/query", body, &results, http.StatusOK); err != nil {
		return nil, fmt.Errorf("reading page template %q: %w", id, err)
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("page template %q not found", id)
	}
	return &results[0], nil
}

func (c *Client) UpdatePageTemplate(ctx context.Context, id string, input PageTemplateInput) (*PageTemplate, error) {
	var pt PageTemplate
	if err := c.doJSON(ctx, http.MethodPatch, fmt.Sprintf("%s/%s", pageTemplateBasePath, id), input, &pt, http.StatusOK); err != nil {
		return nil, fmt.Errorf("updating page template %q: %w", id, err)
	}
	return &pt, nil
}

func (c *Client) DeletePageTemplate(ctx context.Context, id string) error {
	if err := c.doJSON(ctx, http.MethodDelete, fmt.Sprintf("%s/%s", pageTemplateBasePath, id), nil, nil, http.StatusNoContent); err != nil {
		return fmt.Errorf("deleting page template %q: %w", id, err)
	}
	return nil
}
