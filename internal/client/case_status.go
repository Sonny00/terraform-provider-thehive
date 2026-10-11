package client

import (
	"context"
	"fmt"
	"net/http"
)

type CaseStatusInput struct {
	Value string `json:"value"`
	Stage string `json:"stage"`
	Color string `json:"colour,omitempty"`
}

type CaseStatusUpdateInput struct {
	Color string `json:"colour,omitempty"`
}

type CaseStatus struct {
	ID        string `json:"_id"`
	Value     string `json:"value"`
	Stage     string `json:"stage"`
	Color     string `json:"colour"`
	CreatedAt int64  `json:"_createdAt,omitempty"`
	CreatedBy string `json:"_createdBy,omitempty"`
	UpdatedAt *int64 `json:"_updatedAt,omitempty"`
	UpdatedBy string `json:"_updatedBy,omitempty"`
}

const caseStatusBasePath = "/api/v1/caseStatus"

func (c *Client) CreateCaseStatus(ctx context.Context, input CaseStatusInput) (*CaseStatus, error) {
	var status CaseStatus
	if err := c.doJSON(ctx, http.MethodPost, caseStatusBasePath, input, &status, http.StatusCreated); err != nil {
		return nil, fmt.Errorf("creating case status: %w", err)
	}
	return &status, nil
}

func (c *Client) GetCaseStatus(ctx context.Context, id string) (*CaseStatus, error) {
	body := map[string]any{
		"query": []map[string]any{
			{"_name": "getCaseStatus", "idOrName": id},
		},
	}
	var results []CaseStatus
	if err := c.doJSON(ctx, http.MethodPost, "/api/v1/query", body, &results, http.StatusOK); err != nil {
		return nil, fmt.Errorf("reading case status %q: %w", id, err)
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("case status %q not found", id)
	}
	return &results[0], nil
}

func (c *Client) UpdateCaseStatus(ctx context.Context, id string, input CaseStatusUpdateInput) (*CaseStatus, error) {
	body, statusCode, err := c.doRequest(ctx, http.MethodPatch, fmt.Sprintf("%s/%s", caseStatusBasePath, id), input)
	if err != nil {
		return nil, fmt.Errorf("updating case status %q: %w", id, err)
	}
	if statusCode < http.StatusOK || statusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("updating case status %q: %w", id, parseAPIError(statusCode, body))
	}
	return c.GetCaseStatus(ctx, id)
}

func (c *Client) DeleteCaseStatus(ctx context.Context, id string) error {
	if err := c.doJSON(ctx, http.MethodDelete, fmt.Sprintf("%s/%s", caseStatusBasePath, id), nil, nil, http.StatusNoContent); err != nil {
		return fmt.Errorf("deleting case status %q: %w", id, err)
	}
	return nil
}
