# PromptJuggler Go SDK

The official Go client for the [PromptJuggler](https://promptjuggler.com) API. Run prompts and
workflows, manage knowledge bases, and verify webhooks — with typed structs and flat,
synchronous methods, over the standard library's `net/http` (no third-party HTTP stack).

## Requirements

- Go 1.22+

## Installation

```bash
go get go.promptjuggler.com/sdk
```

## Usage

```go
package main

import (
	"context"
	"fmt"

	"go.promptjuggler.com/sdk"
	"go.promptjuggler.com/sdk/client"
)

func main() {
	pj := promptjuggler.New("your-api-key")
	ctx := context.Background()

	// Trigger a run (async — returns the run ID)
	created, err := pj.RunPrompt(ctx, "greeting", "production", map[string]string{"name": "Ada"})
	if err != nil {
		panic(err)
	}

	// Poll for the result
	run, err := pj.GetPromptRun(ctx, created.Id)
	if err != nil {
		panic(err)
	}
	if run.Status == client.RUNSTATUS_COMPLETED {
		fmt.Println(run.GetOutput())
	}
}
```

Errors surface as `*APIError` (with a `StatusCode`), `*NetworkError` (no response), or
`*DecodeError` (an undecodable success body), matched with `errors.As`. Verify incoming webhooks
with `VerifyWebhookSignature`.

## Documentation

Full guides and the API reference: **https://docs.promptjuggler.com/sdks/go/overview**

## License

MIT
