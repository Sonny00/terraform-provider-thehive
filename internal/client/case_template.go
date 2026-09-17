package client

import (
	"context"
	"fmt"
	"net/http"
)

type CaseTemplate struct {
	ID          string   `json:"_id"`
	Name        string   `json:"name"`
	DisplayName string   `json:"displayName"`
	TitlePrefix string   `json:"titlePrefix"`
	Description string   `json:"description"`
	Severity    int      `json:"severity"`
	Tags        []string `json:"tags,omitempty"`
	TLP         int      `json:"tlp"`
	PAP         int      `json:"pap"`
	Summary     string   `json:"summary"`
}

type CaseTemplateInput struct {
	Name        string   `json:"name"`
	DisplayName string   `json:"displayName,omitempty"`
	TitlePrefix string   `json:"titlePrefix,omitempty"`
	Description string   `json:"description,omitempty"`
	Severity    int      `json:"severity,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	TLP         int      `json:"tlp,omitempty"`
	PAP         int      `json:"pap,omitempty"`
	Summary     string   `json:"summary,omitempty"`
}

const caseTemplateBasePath = "/api/v1/caseTemplate"

func (c *Client) CreateCaseTemplate(ctx context.Context, input CaseTemplateInput) (*CaseTemplate, error) {
	var ct CaseTemplate
	if err := c.doJSON(ctx, http.MethodPost, caseTemplateBasePath, input, &ct, http.StatusCreated); err != nil {
		return nil, fmt.Errorf("creating case template: %w", err)
	}
	return &ct, nil
}

func (c *Client) GetCaseTemplate(ctx context.Context, id string) (*CaseTemplate, error) {
	body := map[string]any{
		"query": []map[string]any{
			{"_name": "getCaseTemplate", "idOrName": id},
		},
	}
	var results []CaseTemplate
	if err := c.doJSON(ctx, http.MethodPost, "/api/v1/query", body, &results, http.StatusOK); err != nil {
		return nil, fmt.Errorf("reading case template %q: %w", id, err)
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("case template %q not found", id)
	}
	return &results[0], nil
}

func (c *Client) UpdateCaseTemplate(ctx context.Context, id string, input CaseTemplateInput) (*CaseTemplate, error) {
	var ct CaseTemplate
	if err := c.doJSON(ctx, http.MethodPatch, fmt.Sprintf("%s/%s", caseTemplateBasePath, id), input, &ct, http.StatusOK); err != nil {
		return nil, fmt.Errorf("updating case template %q: %w", id, err)
	}
	return &ct, nil
}

func (c *Client) DeleteCaseTemplate(ctx context.Context, id string) error {
	if err := c.doJSON(ctx, http.MethodDelete, fmt.Sprintf("%s/%s", caseTemplateBasePath, id), nil, nil, http.StatusNoContent); err != nil {
		return fmt.Errorf("deleting case template %q: %w", id, err)
	}
	return nil
}
