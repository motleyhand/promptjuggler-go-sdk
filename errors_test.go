package promptjuggler_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	promptjuggler "go.promptjuggler.com/sdk"
)

func TestNon2xxBecomesAPIError(t *testing.T) {
	m := newMockServer(http.StatusNotFound, `{"error":"Prompt run not found"}`)
	defer m.close()

	_, err := m.client().GetPromptRun(context.Background(), uuid1)
	var apiErr *promptjuggler.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %T (%v), want *APIError", err, err)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %d", apiErr.StatusCode)
	}
	if apiErr.Message != "Prompt run not found" {
		t.Errorf("Message = %q", apiErr.Message)
	}
}

func TestServerErrorBecomesAPIError(t *testing.T) {
	m := newMockServer(http.StatusInternalServerError, `{"error":"boom"}`)
	defer m.close()

	_, err := m.client().GetPromptRun(context.Background(), uuid1)
	var apiErr *promptjuggler.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusInternalServerError {
		t.Fatalf("error = %v, want *APIError(500)", err)
	}
}

func TestConnectionFailureBecomesNetworkError(t *testing.T) {
	// Nothing is listening on port 1 — the request never gets a response.
	c := promptjuggler.New("test-key", promptjuggler.WithBaseURL("http://127.0.0.1:1"))

	_, err := c.GetPromptRun(context.Background(), uuid1)
	var netErr *promptjuggler.NetworkError
	if !errors.As(err, &netErr) {
		t.Fatalf("error = %T (%v), want *NetworkError", err, err)
	}
	if netErr.Unwrap() == nil {
		t.Errorf("NetworkError should wrap the underlying cause")
	}
}

func TestUndecodable2xxBecomesDecodeError(t *testing.T) {
	// A 200 whose body is missing required PromptRevision fields: the generated client decodes
	// it as an error, which must surface as a DecodeError — not a bogus APIError{200}.
	m := newMockServer(http.StatusOK, `{}`)
	defer m.close()

	_, err := m.client().GetPrompt(context.Background(), "greeting", "production")
	var decErr *promptjuggler.DecodeError
	if !errors.As(err, &decErr) {
		t.Fatalf("error = %T (%v), want *DecodeError", err, err)
	}
	if decErr.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want 200", decErr.StatusCode)
	}
	if decErr.Unwrap() == nil {
		t.Errorf("DecodeError should wrap the underlying decode failure")
	}
}
