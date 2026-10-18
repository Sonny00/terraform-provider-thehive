package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type AlertFeederInterval struct {
	Value int64  `json:"value"`
	Unit  string `json:"unit"`
}

type AlertFeederHeader struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type AlertFeederAuth struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Type     string `json:"type"`
}

type AlertFeederTimeout struct {
	Value int64  `json:"value"`
	Unit  string `json:"unit"`
}

type AlertFeederInput struct {
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Method          string                 `json:"method"`
	URL             string                 `json:"url"`
	Interval        AlertFeederInterval    `json:"interval"`
	FunctionName    string                 `json:"functionName"`
	Body            map[string]interface{} `json:"body,omitempty"`
	Headers         []AlertFeederHeader    `json:"headers,omitempty"`
	Enabled         bool                   `json:"enabled"`
	Auth            *AlertFeederAuth       `json:"auth,omitempty"`
	RequestTimeout  *AlertFeederTimeout    `json:"requestTimeout,omitempty"`
	ResponseMaxSize *int64                 `json:"responseMaxSize,omitempty"`
}

type AlertFeederFunction struct {
	ID   string `json:"_id"`
	Name string `json:"name"`
}

type AlertFeeder struct {
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Method          string                 `json:"method"`
	URL             string                 `json:"url"`
	Interval        AlertFeederInterval    `json:"interval"`
	Function        AlertFeederFunction    `json:"function"`
	Body            map[string]interface{} `json:"body,omitempty"`
	Headers         []AlertFeederHeader    `json:"headers,omitempty"`
	Enabled         bool                   `json:"enabled"`
	Auth            *AlertFeederAuth       `json:"auth,omitempty"`
	RequestTimeout  *AlertFeederTimeout    `json:"requestTimeout,omitempty"`
	ResponseMaxSize *int64                 `json:"responseMaxSize,omitempty"`
}

const alertFeederBasePath = "/api/v1/connector/alert-feeder"

func (c *Client) CreateAlertFeeder(ctx context.Context, input AlertFeederInput) (*AlertFeeder, error) {
	body, status, err := c.doRequest(ctx, http.MethodPost, alertFeederBasePath, input)
	if err != nil {
		return nil, fmt.Errorf("creating alert feeder: %w", err)
	}
	if status < http.StatusOK || status >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("creating alert feeder: %w", parseAPIError(status, body))
	}
	return &AlertFeeder{Name: input.Name}, nil
}

func (c *Client) GetAlertFeeder(ctx context.Context, name string) (*AlertFeeder, error) {
	var feeders []AlertFeeder
	if err := c.doJSON(ctx, http.MethodGet, alertFeederBasePath, nil, &feeders, http.StatusOK); err != nil {
		return nil, fmt.Errorf("listing alert feeders: %w", err)
	}
	for i := range feeders {
		if feeders[i].Name == name {
			return &feeders[i], nil
		}
	}
	return nil, fmt.Errorf("alert feeder %q not found", name)
}

func (c *Client) UpdateAlertFeeder(ctx context.Context, name string, input AlertFeederInput) (*AlertFeeder, error) {
	body, status, err := c.doRequest(ctx, http.MethodPut, fmt.Sprintf("%s/%s", alertFeederBasePath, url.PathEscape(name)), input)
	if err != nil {
		return nil, fmt.Errorf("updating alert feeder %q: %w", name, err)
	}
	if status < http.StatusOK || status >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("updating alert feeder %q: %w", name, parseAPIError(status, body))
	}
	updatedName := input.Name
	if updatedName == "" {
		updatedName = name
	}
	return c.GetAlertFeeder(ctx, updatedName)
}

func (c *Client) DeleteAlertFeeder(ctx context.Context, name string) error {
	if err := c.doJSON(ctx, http.MethodDelete, fmt.Sprintf("%s/%s", alertFeederBasePath, url.PathEscape(name)), nil, nil, http.StatusNoContent); err != nil {
		return fmt.Errorf("deleting alert feeder %q: %w", name, err)
	}
	return nil
}
