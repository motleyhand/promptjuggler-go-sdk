package promptjuggler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"go.promptjuggler.com/sdk/client"
)

// APIError is returned when the API responds with a non-2xx status. It carries the HTTP status
// code and the server's error message.
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("promptjuggler: API error %d: %s", e.StatusCode, e.Message)
}

// NetworkError is returned when the request never reached the API (DNS failure, timeout, offline).
type NetworkError struct {
	Err error
}

func (e *NetworkError) Error() string {
	return fmt.Sprintf("promptjuggler: network error: %v", e.Err)
}

func (e *NetworkError) Unwrap() error { return e.Err }

// DecodeError is returned when the API replied with a success status but the SDK could not decode
// its body into the expected model — a missing required field or a mistyped value, typically. The
// request succeeded; the client just couldn't read the response.
type DecodeError struct {
	StatusCode int
	Err        error
}

func (e *DecodeError) Error() string {
	return fmt.Sprintf("promptjuggler: failed to decode the %d response: %v", e.StatusCode, e.Err)
}

func (e *DecodeError) Unwrap() error { return e.Err }

// translate converts the generated client's (response, error) pair into an *APIError (the API
// answered with a non-2xx), a *NetworkError (the request never got a response), or a *DecodeError
// (a success status whose body couldn't be decoded).
func translate(resp *http.Response, err error) error {
	if err == nil {
		return nil
	}
	if resp == nil {
		return &NetworkError{Err: err}
	}
	if resp.StatusCode < 300 {
		// A 2xx whose body the client couldn't decode — not an API error; the request itself
		// succeeded. Surface the underlying decode failure rather than a bogus APIError{200}.
		return &DecodeError{StatusCode: resp.StatusCode, Err: err}
	}
	msg := resp.Status
	var apiErr *client.GenericOpenAPIError
	if errors.As(err, &apiErr) {
		if m := parseErrorMessage(apiErr.Body()); m != "" {
			msg = m
		}
	}
	return &APIError{StatusCode: resp.StatusCode, Message: msg}
}

// parseErrorMessage pulls {"error": "..."} out of a JSON error body, or returns "".
func parseErrorMessage(body []byte) string {
	var parsed struct {
		Error string `json:"error"`
	}
	if json.Unmarshal(body, &parsed) == nil {
		return parsed.Error
	}
	return ""
}
