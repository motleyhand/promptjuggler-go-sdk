package promptjuggler_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func tempFile(t *testing.T, name, content string) *os.File {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open temp file: %v", err)
	}
	t.Cleanup(func() { _ = f.Close() })
	return f
}

func TestUploadDocumentsSendsBracketIndexedParts(t *testing.T) {
	docs := `[{"id":"` + uuid1 + `","status":"pending","fileName":"a.txt","bytes":1,"chunkCount":0},` +
		`{"id":"` + uuid2 + `","status":"pending","fileName":"b.txt","bytes":1,"chunkCount":0}]`
	m := newMockServer(200, docs)
	defer m.close()

	result, err := m.client().UploadDocuments(context.Background(), "product-docs",
		tempFile(t, "a.txt", "alpha"), tempFile(t, "b.txt", "beta"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("got %d docs, want 2", len(result))
	}

	req := m.first()
	if req.path != "/api/v1/knowledge-bases/product-docs/documents" {
		t.Errorf("path = %q", req.path)
	}
	if got := req.header.Get("Authorization"); got != "Bearer test-key" {
		t.Errorf("Authorization = %q", got)
	}
	if !strings.HasPrefix(req.header.Get("Content-Type"), "multipart/form-data") {
		t.Errorf("Content-Type = %q", req.header.Get("Content-Type"))
	}
	body := string(req.body)
	for _, want := range []string{`name="files[0]"`, `name="files[1]"`, `filename="a.txt"`, `filename="b.txt"`} {
		if !strings.Contains(body, want) {
			t.Errorf("multipart body missing %q", want)
		}
	}
}
