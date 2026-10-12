package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type EmailIntakeBasicAuth struct {
	Password string `json:"password,omitempty"`
}

type EmailIntakeOAuth2 struct {
	ClientID          string   `json:"clientId,omitempty"`
	TenantID          string   `json:"tenantId,omitempty"`
	Secret            string   `json:"secret,omitempty"`
	Authority         string   `json:"authority,omitempty"`
	Scopes            []string `json:"scopes,omitempty"`
	RedirectURI       string   `json:"redirectUri,omitempty"`
	AuthorizationCode string   `json:"authorizationCode,omitempty"`
}

type EmailIntakeCredential struct {
	Email     string                `json:"email"`
	BasicAuth *EmailIntakeBasicAuth `json:"basicAuth,omitempty"`
	OAuth2    *EmailIntakeOAuth2    `json:"oAuth2,omitempty"`
}

type EmailIntakeProvider struct {
	Name string `json:"name"`
}

type EmailIntakeMailbox struct {
	Provider   EmailIntakeProvider   `json:"provider"`
	Credential EmailIntakeCredential `json:"credential"`
	Inbox      string                `json:"inbox,omitempty"`
	Archive    string                `json:"archive,omitempty"`
	MarkAsRead bool                  `json:"markAsRead"`
	Kind       string                `json:"_kind,omitempty"`
}

type EmailIntakeAlertProperties struct {
	Type   string   `json:"type,omitempty"`
	Source string   `json:"source,omitempty"`
	Tags   []string `json:"tags,omitempty"`
}

type EmailIntakeInput struct {
	Name            string                      `json:"name"`
	Mailbox         EmailIntakeMailbox          `json:"mailbox"`
	Organisations   []string                    `json:"organisations"`
	Enabled         bool                        `json:"enabled"`
	AlertProperties *EmailIntakeAlertProperties `json:"alertProperties,omitempty"`
}

type EmailIntake struct {
	ID              string                      `json:"id"`
	Name            string                      `json:"name"`
	Mailbox         EmailIntakeMailbox          `json:"mailbox"`
	Organisations   []string                    `json:"organisations"`
	Enabled         bool                        `json:"enabled"`
	AlertProperties *EmailIntakeAlertProperties `json:"alertProperties,omitempty"`
	CreatedAt       int64                       `json:"createdAt,omitempty"`
}

const emailIntakeBasePath = "/api/v1/connector/email-intake/config"

func (c *Client) CreateEmailIntake(ctx context.Context, input EmailIntakeInput) (*EmailIntake, error) {
	var result EmailIntake
	if err := c.doJSON(ctx, http.MethodPost, emailIntakeBasePath, input, &result, http.StatusCreated); err != nil {
		return nil, fmt.Errorf("creating email intake: %w", err)
	}
	return &result, nil
}

func (c *Client) GetEmailIntake(ctx context.Context, id string) (*EmailIntake, error) {
	var result EmailIntake
	if err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("%s/%s", emailIntakeBasePath, id), nil, &result, http.StatusOK); err != nil {
		return nil, fmt.Errorf("getting email intake %q: %w", id, err)
	}
	return &result, nil
}

func (c *Client) UpdateEmailIntake(ctx context.Context, id string, input EmailIntakeInput) (*EmailIntake, error) {
	body, status, err := c.doRequest(ctx, http.MethodPut, fmt.Sprintf("%s/%s", emailIntakeBasePath, id), input)
	if err != nil {
		return nil, fmt.Errorf("updating email intake %q: %w", id, err)
	}
	if status < http.StatusOK || status >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("updating email intake %q: %w", id, parseAPIError(status, body))
	}
	if len(body) == 0 {
		return c.GetEmailIntake(ctx, id)
	}

	var result EmailIntake
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decoding updated email intake %q: %w (body: %s)", id, err, string(body))
	}
	return &result, nil
}

func (c *Client) DeleteEmailIntake(ctx context.Context, id string) error {
	if err := c.doJSON(ctx, http.MethodDelete, fmt.Sprintf("%s/%s", emailIntakeBasePath, id), nil, nil, http.StatusNoContent); err != nil {
		return fmt.Errorf("deleting email intake %q: %w", id, err)
	}
	return nil
}
