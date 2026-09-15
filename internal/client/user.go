package client

import (
	"context"
	"fmt"
	"net/http"
)

type User struct {
	ID           string   `json:"_id"`
	Login        string   `json:"login"`
	Name         string   `json:"name"`
	Profile      string   `json:"profile"`
	Organisation string   `json:"organisation,omitempty"`
	Email        string   `json:"email,omitempty"`
	Type         string   `json:"type,omitempty"`
	Locked       bool     `json:"locked"`
	HasKey       bool     `json:"hasKey"`
}

type UserInput struct {
	Login        string `json:"login"`
	Name         string `json:"name"`
	Profile      string `json:"profile"`
	Organisation string `json:"organisation,omitempty"`
	Email        string `json:"email,omitempty"`
	Type         string `json:"type,omitempty"`
}

type UserUpdateInput struct {
	Name    string `json:"name,omitempty"`
	Profile string `json:"profile,omitempty"`
	Email   string `json:"email,omitempty"`
	Locked  *bool  `json:"locked,omitempty"`
}

const userBasePath = "/api/v1/user"

func (c *Client) CreateUser(ctx context.Context, input UserInput) (*User, error) {
	var u User
	if err := c.doJSON(ctx, http.MethodPost, userBasePath, input, &u, http.StatusCreated); err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}
	return &u, nil
}

func (c *Client) GetUser(ctx context.Context, id string) (*User, error) {
	var u User
	if err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("%s/%s", userBasePath, id), nil, &u, http.StatusOK); err != nil {
		return nil, fmt.Errorf("reading user %q: %w", id, err)
	}
	return &u, nil
}

func (c *Client) UpdateUser(ctx context.Context, id string, input UserUpdateInput) (*User, error) {
	var u User
	if err := c.doJSON(ctx, http.MethodPatch, fmt.Sprintf("%s/%s", userBasePath, id), input, &u, http.StatusOK); err != nil {
		return nil, fmt.Errorf("updating user %q: %w", id, err)
	}
	return &u, nil
}

func (c *Client) DeleteUser(ctx context.Context, id string) error {
	if err := c.doJSON(ctx, http.MethodDelete, fmt.Sprintf("%s/%s", userBasePath, id), nil, nil, http.StatusNoContent); err != nil {
		return fmt.Errorf("deleting user %q: %w", id, err)
	}
	return nil
}

func (c *Client) SetUserPassword(ctx context.Context, id string, password string) error {
	body := map[string]string{"password": password}
	if err := c.doJSON(ctx, http.MethodPost, fmt.Sprintf("%s/%s/password/set", userBasePath, id), body, nil, http.StatusNoContent); err != nil {
		return fmt.Errorf("setting password for user %q: %w", id, err)
	}
	return nil
}

func (c *Client) LockUser(ctx context.Context, id string) error {
	if err := c.doJSON(ctx, http.MethodDelete, fmt.Sprintf("%s/%s", userBasePath, id), nil, nil, http.StatusNoContent); err != nil {
		return fmt.Errorf("locking user %q: %w", id, err)
	}
	return nil
}
