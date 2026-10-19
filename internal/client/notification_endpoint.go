package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type NotificationEndpointAuth struct {
	Type     string `json:"type"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type NotificationEndpointInput struct {
	Name                         string                    `json:"name"`
	Type                         string                    `json:"type"`
	URL                          string                    `json:"url"`
	Version                      int64                     `json:"version"`
	Auth                         *NotificationEndpointAuth `json:"auth,omitempty"`
	IncludedTheHiveOrganisations []string                  `json:"includedTheHiveOrganisations"`
	ExcludedTheHiveOrganisations []string                  `json:"excludedTheHiveOrganisations"`
	WsConfig                     map[string]interface{}    `json:"wsConfig,omitempty"`
}

type NotificationEndpoint struct {
	Name                         string                    `json:"name"`
	Type                         string                    `json:"type"`
	URL                          string                    `json:"url"`
	Version                      int64                     `json:"version"`
	Auth                         *NotificationEndpointAuth `json:"auth,omitempty"`
	IncludedTheHiveOrganisations []string                  `json:"includedTheHiveOrganisations"`
	ExcludedTheHiveOrganisations []string                  `json:"excludedTheHiveOrganisations"`
}

type NotificationEndpointList struct {
	Global       []NotificationEndpoint `json:"global"`
	Organisation []NotificationEndpoint `json:"organisation"`
}

const notificationEndpointPath = "/api/v1/admin/config/organisation/endpoint"
const notificationEndpointListPath = "/api/v1/admin/config/endpoint"

func (c *Client) CreateOrUpdateEndpoint(ctx context.Context, input NotificationEndpointInput) error {
	body, status, err := c.doRequest(ctx, http.MethodPut, notificationEndpointPath, input)
	if err != nil {
		return fmt.Errorf("creating/updating notification endpoint: %w", err)
	}
	if status < http.StatusOK || status >= http.StatusMultipleChoices {
		return fmt.Errorf("creating/updating notification endpoint: %w", parseAPIError(status, body))
	}
	return nil
}

func (c *Client) GetEndpoint(ctx context.Context, name string) (*NotificationEndpoint, error) {
	var result NotificationEndpointList
	if err := c.doJSON(ctx, http.MethodGet, notificationEndpointListPath, nil, &result, http.StatusOK); err != nil {
		return nil, fmt.Errorf("listing notification endpoints: %w", err)
	}

	for i := range result.Organisation {
		if result.Organisation[i].Name == name {
			return &result.Organisation[i], nil
		}
	}

	return nil, fmt.Errorf("notification endpoint %q not found", name)
}

func (c *Client) DeleteEndpoint(ctx context.Context, name string) error {
	path := fmt.Sprintf("%s/%s", notificationEndpointPath, url.PathEscape(name))
	if err := c.doJSON(ctx, http.MethodDelete, path, nil, nil, http.StatusNoContent); err != nil {
		return fmt.Errorf("deleting notification endpoint %q: %w", name, err)
	}
	return nil
}
