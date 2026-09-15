package client

import (
	"context"
	"fmt"
	"net/http"
)

// ---------- Models ----------

// Organisation is the API representation returned by TheHive.
//
// NOTE: field names follow the TheHive 5 API response schema.
// Verify against your instance's Swagger/OpenAPI spec if they drift.
type Organisation struct {
	ID          string `json:"_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// OrganisationInput is the payload sent for create/update.
type OrganisationInput struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// ---------- CRUD ----------

const orgBasePath = "/api/v1/organisation"

// CreateOrganisation creates a new organisation and returns it.
func (c *Client) CreateOrganisation(ctx context.Context, input OrganisationInput) (*Organisation, error) {
	var org Organisation
	if err := c.doJSON(ctx, http.MethodPost, orgBasePath, input, &org, http.StatusCreated); err != nil {
		return nil, fmt.Errorf("creating organisation: %w", err)
	}
	return &org, nil
}

// GetOrganisation fetches an organisation by its ID.
func (c *Client) GetOrganisation(ctx context.Context, id string) (*Organisation, error) {
	var org Organisation
	path := fmt.Sprintf("%s/%s", orgBasePath, id)
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &org, http.StatusOK); err != nil {
		return nil, fmt.Errorf("reading organisation %q: %w", id, err)
	}
	return &org, nil
}

// UpdateOrganisation patches an existing organisation and returns the updated state.
func (c *Client) UpdateOrganisation(ctx context.Context, id string, input OrganisationInput) (*Organisation, error) {
	var org Organisation
	path := fmt.Sprintf("%s/%s", orgBasePath, id)
	if err := c.doJSON(ctx, http.MethodPatch, path, input, &org, http.StatusOK); err != nil {
		return nil, fmt.Errorf("updating organisation %q: %w", id, err)
	}
	return &org, nil
}

// DeleteOrganisation removes an organisation.
func (c *Client) DeleteOrganisation(ctx context.Context, id string) error {
	path := fmt.Sprintf("%s/%s", orgBasePath, id)
	if err := c.doJSON(ctx, http.MethodDelete, path, nil, nil, http.StatusNoContent); err != nil {
		return fmt.Errorf("deleting organisation %q: %w", id, err)
	}
	return nil
}
