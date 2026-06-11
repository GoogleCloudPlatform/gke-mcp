---
name: gke-mcp-tool-generator
description: Dynamically generates Go boilerplate and registration code for new GKE MCP tools.
---

# GKE MCP Tool Generator

Use this skill to automate the creation of new Go-based MCP tools in the `gke-mcp` repository.

## Workflow

### 1. Gather Input
Ask the user for the following details if not already provided:
*   **Tool Name**: The name of the tool (e.g., `list_subnets`). It should use `snake_case`.
*   **Description**: A clear description of what the tool does (this will be exposed to the LLM).
*   **Arguments**: A list of arguments, including their types, descriptions, and whether they are required.

### 2. Generate Implementation File
Create the implementation file at `pkg/tools/<tool-package-name>/<tool-package-name>.go` (where `<tool-package-name>` is the tool name without underscores or a simplified kebab/snake version, e.g., `listsubnets` or `subnets`).
By convention, package names in Go should be lowercase, single-word if possible.

Generate the code using this template:

```go
package <package-name>

import (
	"context"
	"fmt"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// <ToolArgsStructName> defines the arguments for the <tool_name> tool.
type <ToolArgsStructName> struct {
	// Add arguments here with json and jsonschema tags.
	// Example:
	// ProjectID string `json:"project_id" jsonschema:"Required. The GCP project ID."`
}

type handlers struct {
	c *config.Config
	// Add clients here (e.g., GKE client)
}

// Install registers the tool with the MCP server.
func Install(ctx context.Context, s *mcp.Server, c *config.Config) error {
	// Initialize clients if needed
	h := &handlers{
		c: c,
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "<tool_name>",
		Description: "<tool_description>",
	}, h.<handlerMethodName>)

	return nil
}

func (h *handlers) <handlerMethodName>(ctx context.Context, _ *mcp.CallToolRequest, args *<ToolArgsStructName>) (*mcp.CallToolResult, any, error) {
	// TODO: Implement tool logic

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Tool <tool_name> invoked with args: %+v", args),
			},
		},
	}, nil, nil
}
```

### 3. Generate Test File
Create a basic unit test file at `pkg/tools/<tool-package-name>/<tool-package-name>_test.go`:

```go
package <package-name>

import (
	"context"
	"testing"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func Test<HandlerMethodName>(t *testing.T) {
	ctx := context.Background()
	c := config.New("test-version", false)
	h := &handlers{
		c: c,
	}

	args := &<ToolArgsStructName>{
		// Initialize test arguments
	}

	resp, _, err := h.<handlerMethodName>(ctx, &mcp.CallToolRequest{}, args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Content) == 0 {
		t.Fatal("expected response content, got none")
	}
	
	// Add assertions
}
```

### 4. Register in Main Server
Update [pkg/tools/tools.go](../../pkg/tools/tools.go) to register the new tool:
1.  Add the import:
    ```go
    "github.com/GoogleCloudPlatform/gke-mcp/pkg/tools/<package-name>"
    ```
2.  Add the installer to the `installers` list in `Install` function:
    ```go
    installers := []installer{
        // ...
        <package-name>.Install,
    }
    ```

## Execution
If you have file-writing capabilities:
1.  Create the directory `pkg/tools/<package-name>/`.
2.  Write the implementation file.
3.  Write the test file.
4.  Update `pkg/tools/tools.go` with the new import and registry entry.
5.  Run `go test ./pkg/tools/<package-name>/...` to ensure it compiles and tests pass.
6.  Inform the user that the skeleton has been created and verified.
