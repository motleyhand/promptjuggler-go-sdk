package promptjuggler_test

import (
	"io"
	"net/http"
	"net/http/httptest"

	promptjuggler "go.promptjuggler.com/sdk"
)

// capturedRequest is one request the mock server received.
type capturedRequest struct {
	method string
	path   string
	header http.Header
	body   []byte
}

// mockServer is a real local HTTP server (httptest) that records every request and replies with
// a fixed status + body — the Go analogue of the other SDKs' mock transports.
type mockServer struct {
	server   *httptest.Server
	captured []capturedRequest
}

func newMockServer(status int, body string) *mockServer {
	m := &mockServer{}
	m.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		m.captured = append(m.captured, capturedRequest{r.Method, r.URL.Path, r.Header.Clone(), b})
		// The generated client only decodes responses typed application/json.
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = io.WriteString(w, body)
	}))
	return m
}

func (m *mockServer) client() *promptjuggler.Client {
	return promptjuggler.New("test-key", promptjuggler.WithBaseURL(m.server.URL))
}

func (m *mockServer) first() capturedRequest { return m.captured[0] }

func (m *mockServer) close() { m.server.Close() }
