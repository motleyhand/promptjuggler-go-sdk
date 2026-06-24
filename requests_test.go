package promptjuggler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	promptjuggler "go.promptjuggler.com/sdk"
	"go.promptjuggler.com/sdk/client"
)

const (
	uuid1 = "550e8400-e29b-41d4-a716-446655440000"
	uuid2 = "550e8400-e29b-41d4-a716-446655440001"
)

var (
	revisionJSON = `{"id":"` + uuid1 + `","promptId":"` + uuid2 + `","memory":"stateless",` +
		`"provider":"openai","model":"gpt-4o","modelParams":{},"responseFormat":{"type":"text"},` +
		`"messages":[],"tools":[]}`
	runResponseJSON = `{"id":"` + uuid1 + `","thread":"` + uuid2 + `"}`
)

func body(t *testing.T, req capturedRequest) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(req.body, &m); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	return m
}

func TestGetPromptSendsGetWithBearer(t *testing.T) {
	m := newMockServer(http.StatusOK, revisionJSON)
	defer m.close()
	if _, err := m.client().GetPrompt(context.Background(), "greeting", "production"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	req := m.first()
	if req.method != http.MethodGet {
		t.Errorf("method = %q, want GET", req.method)
	}
	if req.path != "/api/v1/prompts/greeting/production" {
		t.Errorf("path = %q", req.path)
	}
	if got := req.header.Get("Authorization"); got != "Bearer test-key" {
		t.Errorf("Authorization = %q", got)
	}
}

func TestGetPromptAcceptsNumericVersion(t *testing.T) {
	m := newMockServer(http.StatusOK, revisionJSON)
	defer m.close()
	if _, err := m.client().GetPrompt(context.Background(), "greeting", "42"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := m.first().path; got != "/api/v1/prompts/greeting/42" {
		t.Errorf("path = %q", got)
	}
}

func TestRunPromptPostsInputsOnly(t *testing.T) {
	m := newMockServer(http.StatusOK, runResponseJSON)
	defer m.close()
	_, err := m.client().RunPrompt(context.Background(), "greeting", "production",
		map[string]string{"name": "Ada"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	req := m.first()
	if req.method != http.MethodPost {
		t.Errorf("method = %q, want POST", req.method)
	}
	if req.path != "/api/v1/prompts/greeting/production/runs" {
		t.Errorf("path = %q", req.path)
	}
	if got := body(t, req)["inputs"].(map[string]any)["name"]; got != "Ada" {
		t.Errorf("inputs.name = %v", got)
	}
}

func TestRunPromptSerializesOptionsAndArrayMetadata(t *testing.T) {
	m := newMockServer(http.StatusOK, runResponseJSON)
	defer m.close()
	_, err := m.client().RunPrompt(context.Background(), "greeting", "1",
		map[string]string{"topic": "AI safety"},
		promptjuggler.WithPriority("onsite"),
		promptjuggler.WithEnvironment("staging"),
		promptjuggler.WithEnvVars(map[string]string{"MY_API_KEY": "sk-x"}),
		promptjuggler.WithMetadata(map[string]any{"tags": []string{"a", "b"}, "user_id": "42"}),
		promptjuggler.WithChannel("support"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	b := body(t, m.first())
	if b["priority"] != "onsite" {
		t.Errorf("priority = %v", b["priority"])
	}
	if b["environment"] != "staging" {
		t.Errorf("environment = %v", b["environment"])
	}
	if got := b["envVars"].(map[string]any)["MY_API_KEY"]; got != "sk-x" {
		t.Errorf("envVars.MY_API_KEY = %v", got)
	}
	if b["channel"] != "support" {
		t.Errorf("channel = %v", b["channel"])
	}
	md := b["metadata"].(map[string]any)
	if md["user_id"] != "42" {
		t.Errorf("metadata.user_id = %v", md["user_id"])
	}
	tags := md["tags"].([]any)
	if len(tags) != 2 || tags[0] != "a" {
		t.Errorf("metadata.tags = %v", tags)
	}
}

func TestGetPromptRunGetsByID(t *testing.T) {
	run := `{"id":"` + uuid1 + `","status":"completed","createdAt":"2026-01-01T00:00:00Z"}`
	m := newMockServer(http.StatusOK, run)
	defer m.close()
	if _, err := m.client().GetPromptRun(context.Background(), uuid1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := m.first().path; got != "/api/v1/promptruns/"+uuid1 {
		t.Errorf("path = %q", got)
	}
}

func TestRunWorkflowPostsToWorkflowRuns(t *testing.T) {
	m := newMockServer(http.StatusOK, runResponseJSON)
	defer m.close()
	_, err := m.client().RunWorkflow(context.Background(), "onboarding", "production",
		map[string]string{"email": "a@b.com"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	req := m.first()
	if req.method != http.MethodPost {
		t.Errorf("method = %q, want POST", req.method)
	}
	if req.path != "/api/v1/workflows/onboarding/production/runs" {
		t.Errorf("path = %q", req.path)
	}
}

func TestGetKnowledgeBaseGetsBySlug(t *testing.T) {
	kb := `{"id":"` + uuid1 + `","status":"ready","documentCount":0,"chunkCount":0,"documents":[]}`
	m := newMockServer(http.StatusOK, kb)
	defer m.close()
	if _, err := m.client().GetKnowledgeBase(context.Background(), "product-docs"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := m.first().path; got != "/api/v1/knowledge-bases/product-docs" {
		t.Errorf("path = %q", got)
	}
}

func TestDeleteKnowledgeDocumentDeletesByID(t *testing.T) {
	m := newMockServer(http.StatusNoContent, "")
	defer m.close()
	if err := m.client().DeleteKnowledgeDocument(context.Background(), uuid1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	req := m.first()
	if req.method != http.MethodDelete {
		t.Errorf("method = %q, want DELETE", req.method)
	}
	if req.path != "/api/v1/knowledge-documents/"+uuid1 {
		t.Errorf("path = %q", req.path)
	}
}

// A prompt's tools reference a revision by number or tag (VersionRef.idOrTag). Go's strict
// encoding/json can't coerce a number into a string, so idOrTag is deliberately left as the
// generated oneOf wrapper — which must accept both a numeric ref and a tag.
func TestVersionRefAcceptsNumericAndStringIdOrTag(t *testing.T) {
	for _, raw := range []string{
		`{"parentId":"` + uuid2 + `","idOrTag":1}`,
		`{"parentId":"` + uuid2 + `","idOrTag":"production"}`,
	} {
		var ref client.VersionRef
		if err := json.Unmarshal([]byte(raw), &ref); err != nil {
			t.Errorf("unmarshal %s: %v", raw, err)
		}
	}
}
