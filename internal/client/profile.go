package client

import (
	"context"
	"fmt"
	"net/http"
)

type Profile struct {
	ID          string   `json:"_id"`
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
	IsAdmin     bool     `json:"forAdmin"`
	Editable    bool     `json:"editable"`
}

type ProfileInput struct {
	Name        string   `json:"name"`
	Permissions []string `json:"permissions,omitempty"`
	IsAdmin     bool     `json:"forAdmin,omitempty"`
}

const profileBasePath = "/api/v1/profile"

func (c *Client) GetProfile(ctx context.Context, id string) (*Profile, error) {
	body := map[string]any{
		"query": []map[string]any{
			{"_name": "getProfile", "idOrName": id},
		},
	}
	var results []Profile
	if err := c.doJSON(ctx, http.MethodPost, "/api/v1/query", body, &results, http.StatusOK); err != nil {
		return nil, fmt.Errorf("reading profile %q: %w", id, err)
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("profile %q not found", id)
	}
	return &results[0], nil
}

func (c *Client) CreateProfile(ctx context.Context, input ProfileInput) (*Profile, error) {
	var p Profile
	if err := c.doJSON(ctx, http.MethodPost, profileBasePath, input, &p, http.StatusCreated); err != nil {
		return nil, fmt.Errorf("creating profile: %w", err)
	}
	return &p, nil
}

func (c *Client) UpdateProfile(ctx context.Context, id string, input ProfileInput) (*Profile, error) {
	var p Profile
	if err := c.doJSON(ctx, http.MethodPatch, fmt.Sprintf("%s/%s", profileBasePath, id), input, &p, http.StatusOK); err != nil {
		return nil, fmt.Errorf("updating profile %q: %w", id, err)
	}
	return &p, nil
}

func (c *Client) DeleteProfile(ctx context.Context, id string) error {
	if err := c.doJSON(ctx, http.MethodDelete, fmt.Sprintf("%s/%s", profileBasePath, id), nil, nil, http.StatusNoContent); err != nil {
		return fmt.Errorf("deleting profile %q: %w", id, err)
	}
	return nil
}
