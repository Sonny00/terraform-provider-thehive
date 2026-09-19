package client

import (
	"context"
	"fmt"
	"net/http"
)

type CortexConfig struct {
	StatusCheckInterval string         `json:"statusCheckInterval"`
	RefreshDelay        string         `json:"refreshDelay"`
	MaxRetryOnError     int            `json:"maxRetryOnError"`
	JobTimeout          string         `json:"jobTimeout"`
	Servers             []CortexServer `json:"servers"`
}

type CortexServer struct {
	Name                         string     `json:"name"`
	URL                          string     `json:"url"`
	Auth                         CortexAuth `json:"auth"`
	IncludedTheHiveOrganisations []string   `json:"includedTheHiveOrganisations"`
	ExcludedTheHiveOrganisations []string   `json:"excludedTheHiveOrganisations"`
	NumberOfConcurrentSubmission int        `json:"numberOfConcurrentSubmission,omitempty"`
	MaxResponseSize              int        `json:"maxResponseSize,omitempty"`
}

type CortexAuth struct {
	Type string `json:"type"`
	Key  string `json:"key"`
}

const cortexConfigPath = "/api/v1/admin/config/cortex"

func (c *Client) GetCortexConfig(ctx context.Context) (*CortexConfig, error) {
	var cfg CortexConfig
	if err := c.doJSON(ctx, http.MethodGet, cortexConfigPath, nil, &cfg, http.StatusOK); err != nil {
		return nil, fmt.Errorf("reading cortex config: %w", err)
	}
	return &cfg, nil
}

func (c *Client) UpdateCortexConfig(ctx context.Context, cfg CortexConfig) (*CortexConfig, error) {
	_, status, err := c.doRequest(ctx, http.MethodPut, cortexConfigPath, cfg)
	if err != nil {
		return nil, fmt.Errorf("updating cortex config: %w", err)
	}
	if status != http.StatusOK && status != http.StatusNoContent {
		return nil, fmt.Errorf("updating cortex config: unexpected status %d", status)
	}
	// PUT returns 204 No Content — read back the config
	return c.GetCortexConfig(ctx)
}
