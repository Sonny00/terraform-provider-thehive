package client

import (
	"context"
	"fmt"
	"net/http"
)

type MISPConfig struct {
	Interval string       `json:"interval"`
	Servers  []MISPServer `json:"servers"`
}

type MISPServer struct {
	Name                         string   `json:"name"`
	URL                          string   `json:"url"`
	Auth                         MISPAuth `json:"auth"`
	Purpose                      string   `json:"purpose,omitempty"`
	IncludedTheHiveOrganisations []string `json:"includedTheHiveOrganisations,omitempty"`
	ExcludedTheHiveOrganisations []string `json:"excludedTheHiveOrganisations,omitempty"`
	MaxAge                       int      `json:"maxAge,omitempty"`
	MaxAttributes                int      `json:"maxAttributes,omitempty"`
	ExcludedOrganisations        []string `json:"excludedOrganisations,omitempty"`
}

type MISPAuth struct {
	Type string `json:"type"`
	Key  string `json:"key"`
}

const mispConfigPath = "/api/v1/admin/config/misp"

func (c *Client) GetMISPConfig(ctx context.Context) (*MISPConfig, error) {
	var cfg MISPConfig
	if err := c.doJSON(ctx, http.MethodGet, mispConfigPath, nil, &cfg, http.StatusOK); err != nil {
		return nil, fmt.Errorf("reading misp config: %w", err)
	}
	return &cfg, nil
}

func (c *Client) UpdateMISPConfig(ctx context.Context, cfg MISPConfig) (*MISPConfig, error) {
	_, status, err := c.doRequest(ctx, http.MethodPut, mispConfigPath, cfg)
	if err != nil {
		return nil, fmt.Errorf("updating misp config: %w", err)
	}
	if status != http.StatusOK && status != http.StatusNoContent {
		return nil, fmt.Errorf("updating misp config: unexpected status %d", status)
	}
	return c.GetMISPConfig(ctx)
}
