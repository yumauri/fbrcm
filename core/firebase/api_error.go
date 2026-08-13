package firebase

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// APIError is the stable, presentation-neutral error returned for non-2xx
// Google API responses.
type APIError struct {
	Service      string
	Operation    string
	StatusCode   int
	Status       string
	RemoteStatus string
	RemoteCode   string
	Message      string
	RetryAfter   time.Duration
}

func (e *APIError) Error() string {
	message := strings.TrimSpace(e.Message)
	if message == "" {
		message = e.Status
	}
	return fmt.Sprintf("%s %s returned %s: %s", e.Service, e.Operation, e.Status, message)
}

func newAPIError(service, operation string, resp *http.Response, body []byte) error {
	err := &APIError{Service: service, Operation: operation, StatusCode: resp.StatusCode, Status: resp.Status}
	var payload struct {
		Error struct {
			Code    json.RawMessage `json:"code"`
			Message string          `json:"message"`
			Status  string          `json:"status"`
		} `json:"error"`
	}
	if json.Unmarshal(body, &payload) == nil {
		err.Message = strings.TrimSpace(payload.Error.Message)
		err.RemoteStatus = strings.TrimSpace(payload.Error.Status)
		err.RemoteCode = strings.Trim(string(payload.Error.Code), `"`)
	}
	if err.Message == "" {
		err.Message = strings.TrimSpace(string(body))
	}
	if value := strings.TrimSpace(resp.Header.Get("Retry-After")); value != "" {
		if seconds, parseErr := strconv.Atoi(value); parseErr == nil && seconds >= 0 {
			err.RetryAfter = time.Duration(seconds) * time.Second
		}
	}
	return err
}

func (e *APIError) Retryable() bool {
	return e.StatusCode == http.StatusRequestTimeout || e.StatusCode == http.StatusTooManyRequests || e.StatusCode >= 500
}
