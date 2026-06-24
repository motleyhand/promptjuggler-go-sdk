package promptjuggler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"go.promptjuggler.com/sdk/client"
)

// UploadDocuments uploads one or more documents to a knowledge base (processed asynchronously).
func (c *Client) UploadDocuments(
	ctx context.Context,
	slug string,
	files ...*os.File,
) ([]client.KnowledgeDocumentResponse, error) {
	// The generated client names every multipart part "files"; the server needs them
	// bracket-indexed (files[i]) to parse them as an array. Build the body ourselves and send
	// it through the configured http client, reusing the base URL + bearer auth.
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	for i, f := range files {
		part, err := mw.CreateFormFile(fmt.Sprintf("files[%d]", i), filepath.Base(f.Name()))
		if err != nil {
			return nil, &NetworkError{Err: err}
		}
		if _, err := io.Copy(part, f); err != nil {
			return nil, &NetworkError{Err: err}
		}
	}
	if err := mw.Close(); err != nil {
		return nil, &NetworkError{Err: err}
	}

	url := strings.TrimRight(c.baseURL, "/") + "/api/v1/knowledge-bases/" + slug + "/documents"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &buf)
	if err != nil {
		return nil, &NetworkError{Err: err}
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", mw.FormDataContentType())

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, &NetworkError{Err: err}
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &NetworkError{Err: err}
	}
	if resp.StatusCode/100 != 2 {
		message := resp.Status
		if m := parseErrorMessage(body); m != "" {
			message = m
		}
		return nil, &APIError{StatusCode: resp.StatusCode, Message: message}
	}

	var docs []client.KnowledgeDocumentResponse
	if err := json.Unmarshal(body, &docs); err != nil {
		return nil, &DecodeError{StatusCode: resp.StatusCode, Err: err}
	}
	return docs, nil
}
