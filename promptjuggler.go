// Package promptjuggler is the official Go client for the PromptJuggler API. It wraps the
// generated client (package client) with a flat, ergonomic surface: parameters in, generated
// typed models out, with API errors translated into *APIError / *NetworkError. Synchronous.
package promptjuggler

import (
	"context"
	"net/http"

	"go.promptjuggler.com/sdk/client"
)

const defaultBaseURL = "https://promptjuggler.com"

// Client is an authenticated PromptJuggler API client. It is safe for concurrent use.
type Client struct {
	api     *client.APIClient
	apiKey  string
	baseURL string
	http    *http.Client
}

// Option configures a Client.
type Option func(*Client)

// WithBaseURL points the client at a custom base URL (e.g. for testing).
func WithBaseURL(baseURL string) Option {
	return func(c *Client) { c.baseURL = baseURL }
}

// WithHTTPClient sets the underlying *http.Client (timeouts, transport, proxy, ...).
func WithHTTPClient(h *http.Client) Option {
	return func(c *Client) { c.http = h }
}

// New constructs a client authenticated with the given API key. The key is sent as a Bearer
// token to https://promptjuggler.com unless WithBaseURL overrides it.
func New(apiKey string, opts ...Option) *Client {
	c := &Client{apiKey: apiKey, baseURL: defaultBaseURL, http: &http.Client{}}
	for _, o := range opts {
		o(c)
	}
	cfg := client.NewConfiguration()
	cfg.Servers = client.ServerConfigurations{{URL: c.baseURL}}
	cfg.DefaultHeader["Authorization"] = "Bearer " + c.apiKey
	cfg.HTTPClient = c.http
	c.api = client.NewAPIClient(cfg)
	return c
}

// GetPrompt fetches a prompt revision by slug and version (a revision number or a tag like
// "production"; a numeric revision is passed as its string form, e.g. "42").
func (c *Client) GetPrompt(
	ctx context.Context,
	slug, version string,
) (*client.PromptRevision, error) {
	rev, resp, err := c.api.PromptsAPI.GetPromptRevision(ctx, slug, version).Execute()
	if e := translate(resp, err); e != nil {
		return nil, e
	}
	return rev, nil
}

// RunPrompt triggers a prompt run (async — returns the run ID; poll GetPromptRun for the result).
func (c *Client) RunPrompt(
	ctx context.Context,
	slug, version string,
	inputs map[string]string,
	opts ...RunOption,
) (*client.CreatePromptRunResponse, error) {
	cfg := applyRunOptions(opts)
	body := client.NewCreatePromptRun(inputs)
	if cfg.priority != nil {
		body.SetPriority(*cfg.priority)
	}
	if cfg.thread != nil {
		body.SetThread(*cfg.thread)
	}
	if cfg.environment != nil {
		body.SetEnvironment(*cfg.environment)
	}
	if cfg.envVars != nil {
		body.SetEnvVars(cfg.envVars)
	}
	if cfg.metadata != nil {
		body.SetMetadata(cfg.metadata)
	}
	if cfg.channel != nil {
		body.SetChannel(*cfg.channel)
	}
	res, resp, err := c.api.PromptRunsAPI.CreatePromptRun(ctx, slug, version).
		CreatePromptRun(*body).
		Execute()
	if e := translate(resp, err); e != nil {
		return nil, e
	}
	return res, nil
}

// GetPromptRun fetches a prompt run by ID.
func (c *Client) GetPromptRun(ctx context.Context, runID string) (*client.PromptRun, error) {
	run, resp, err := c.api.PromptRunsAPI.GetPromptRun(ctx, runID).Execute()
	if e := translate(resp, err); e != nil {
		return nil, e
	}
	return run, nil
}

// RunWorkflow triggers a workflow run (async — returns the run ID; poll GetWorkflowRun for the
// result). WithChannel is ignored.
func (c *Client) RunWorkflow(
	ctx context.Context,
	slug, version string,
	inputs map[string]string,
	opts ...RunOption,
) (*client.CreateWorkflowRunResponse, error) {
	cfg := applyRunOptions(opts)
	body := client.NewCreateWorkflowRun(inputs)
	if cfg.priority != nil {
		body.SetPriority(*cfg.priority)
	}
	if cfg.thread != nil {
		body.SetThread(*cfg.thread)
	}
	if cfg.environment != nil {
		body.SetEnvironment(*cfg.environment)
	}
	if cfg.envVars != nil {
		body.SetEnvVars(cfg.envVars)
	}
	if cfg.metadata != nil {
		body.SetMetadata(cfg.metadata)
	}
	res, resp, err := c.api.WorkflowRunsAPI.CreateWorkflowRun(ctx, slug, version).
		CreateWorkflowRun(*body).
		Execute()
	if e := translate(resp, err); e != nil {
		return nil, e
	}
	return res, nil
}

// GetWorkflowRun fetches a workflow run by ID.
func (c *Client) GetWorkflowRun(ctx context.Context, runID string) (*client.WorkflowRun, error) {
	run, resp, err := c.api.WorkflowRunsAPI.GetWorkflowRun(ctx, runID).Execute()
	if e := translate(resp, err); e != nil {
		return nil, e
	}
	return run, nil
}

// GetKnowledgeBase fetches a knowledge base by slug.
func (c *Client) GetKnowledgeBase(
	ctx context.Context,
	slug string,
) (*client.KnowledgeBaseResponse, error) {
	kb, resp, err := c.api.KnowledgeBasesAPI.PublicGetKnowledgeBase(ctx, slug).Execute()
	if e := translate(resp, err); e != nil {
		return nil, e
	}
	return kb, nil
}

// GetKnowledgeDocument fetches a knowledge document by ID.
func (c *Client) GetKnowledgeDocument(
	ctx context.Context,
	documentID string,
) (*client.KnowledgeDocumentResponse, error) {
	doc, resp, err := c.api.KnowledgeBasesAPI.PublicGetDocument(ctx, documentID).Execute()
	if e := translate(resp, err); e != nil {
		return nil, e
	}
	return doc, nil
}

// DeleteKnowledgeDocument deletes a knowledge document by ID.
func (c *Client) DeleteKnowledgeDocument(ctx context.Context, documentID string) error {
	resp, err := c.api.KnowledgeBasesAPI.PublicDeleteDocument(ctx, documentID).Execute()
	return translate(resp, err)
}
