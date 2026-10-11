package client

import (
	"context"
	"fmt"
	"net/http"
)

type AlertStatusInput struct {
	Value string `json:"value"`
	Stage string `json:"stage"`
	Color string `json:"colour,omitempty"`
}

type AlertStatusUpdateInput struct {
	Color string `json:"colour,omitempty"`
}

type AlertStatus struct {
	ID        string `json:"_id"`
	Value     string `json:"value"`
	Stage     string `json:"stage"`
	Color     string `json:"colour"`
	CreatedAt int64  `json:"_createdAt,omitempty"`
	CreatedBy string `json:"_createdBy,omitempty"`
	UpdatedAt *int64 `json:"_updatedAt,omitempty"`
	UpdatedBy string `json:"_updatedBy,omitempty"`
}

const alertStatusBasePath = "/api/v1/alertStatus"

func (c *Client) CreateAlertStatus(ctx context.Context, input AlertStatusInput) (*AlertStatus, error) {
	var status AlertStatus
	if err := c.doJSON(ctx, http.MethodPost, alertStatusBasePath, input, &status, http.StatusCreated); err != nil {
		return nil, fmt.Errorf("creating alert status: %w", err)
	}
	return &status, nil
}

func (c *Client) GetAlertStatus(ctx context.Context, id string) (*AlertStatus, error) {
	body := map[string]any{
		"query": []map[string]any{
			{"_name": "getAlertStatus", "idOrName": id},
		},
	}
	var results []AlertStatus
	if err := c.doJSON(ctx, http.MethodPost, "/api/v1/query", body, &results, http.StatusOK); err != nil {
		return nil, fmt.Errorf("reading alert status %q: %w", id, err)
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("alert status %q not found", id)
	}
	return &results[0], nil
}

func (c *Client) UpdateAlertStatus(ctx context.Context, id string, input AlertStatusUpdateInput) (*AlertStatus, error) {
	var status AlertStatus
	if err := c.doJSON(ctx, http.MethodPatch, fmt.Sprintf("%s/%s", alertStatusBasePath, id), input, &status, http.StatusOK); err != nil {
		return nil, fmt.Errorf("updating alert status %q: %w", id, err)
	}
	return &status, nil
}

func (c *Client) DeleteAlertStatus(ctx context.Context, id string) error {
	if err := c.doJSON(ctx, http.MethodDelete, fmt.Sprintf("%s/%s", alertStatusBasePath, id), nil, nil, http.StatusNoContent); err != nil {
		return fmt.Errorf("deleting alert status %q: %w", id, err)
	}
	return nil
}
