package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type UserError struct {
	Message string
	Hint    string
}

type APIError struct {
	Error string `json:"error"`
	Hint  string `json:"hint,omitempty"`
}

func (e *UserError) Error() string {
	if e.Hint != "" {
		return fmt.Sprintf("%s\n  → %s", e.Message, e.Hint)
	}
	return e.Message
}

func NewUserError(msg, hint string) *UserError {
	return &UserError{Message: msg, Hint: hint}
}

func HandleHTTPError(resp *http.Response) error {
	switch resp.StatusCode {
	case 400:
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return NewUserError("Bad request", "Failed to read error details")
		}
		defer resp.Body.Close()

		var apiErr APIError
		if err := json.Unmarshal(body, &apiErr); err == nil {
			return NewUserError(apiErr.Error, apiErr.Hint)
		}

		return NewUserError("Bad request", "Check your request parameters")
	case 401:
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return NewUserError("Authentication failed", "Failed to read error details")
		}
		defer resp.Body.Close()

		var apiErr APIError
		if err := json.Unmarshal(body, &apiErr); err == nil {
			return NewUserError(apiErr.Error, apiErr.Hint)
		}
		return NewUserError("Authentication failed", "Check your API key or run 'otaship link'")
	case 404:
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return NewUserError("Resource not found", "Failed to read error details")
		}
		defer resp.Body.Close()

		var apiErr APIError
		if err := json.Unmarshal(body, &apiErr); err == nil {
			return NewUserError(apiErr.Error, apiErr.Hint)
		}
		return NewUserError("Resource not found", "Verify project ID in otaship.json")
	case 500:
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return NewUserError("Server error", "Failed to read error details")
		}
		defer resp.Body.Close()

		var apiErr APIError
		if err := json.Unmarshal(body, &apiErr); err == nil {
			return NewUserError(apiErr.Error, apiErr.Hint)
		}
		return NewUserError("Server error", "Contact your OTAShip server admin")
	default:
		return NewUserError(
			fmt.Sprintf("HTTP %d error", resp.StatusCode),
			"Check server logs for details",
		)
	}
}
