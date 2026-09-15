package client

import (
	"encoding/json"
	"fmt"
)

// APIError represents an error returned by the TheHive API.
type APIError struct {
	StatusCode int
	Type       string `json:"type"`
	Message    string `json:"message"`
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("TheHive API %d: %s — %s", e.StatusCode, e.Type, e.Message)
	}
	return fmt.Sprintf("TheHive API %d: unexpected error", e.StatusCode)
}

func parseAPIError(statusCode int, body []byte) error {
	apiErr := &APIError{StatusCode: statusCode}

	if err := json.Unmarshal(body, apiErr); err != nil {
		apiErr.Message = string(body)
	}

	return apiErr
}
