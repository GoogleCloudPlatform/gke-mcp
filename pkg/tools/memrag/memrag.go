// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package memrag

import (
	"context"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	sampleContext = `
Here is how you can complete the requested tasks using the Gemini CLI with GIQ:

1. Which models have been benchmarked by GIQ?
```sh
gcloud alpha container ai profiles models list
```
`
)

type handlers struct {
	c *config.Config
}

func Install(s *server.MCPServer, c *config.Config) {
	h := &handlers{
		c: c,
	}

	memorizeTool := mcp.NewTool("memrag_memorize",
		mcp.WithDescription("Stores or 'memorizes' a piece of text-based context into a long-term knowledge base (memory RAG). This allows the information to be retrieved later using the 'query' tool. Use this to add new information or context that should be remembered."),
		mcp.WithString("context", mcp.Required(), mcp.Description("The context to memorize.")),
	)
	s.AddTool(memorizeTool, h.memorize)

	queryTool := mcp.NewTool("memrag_query",
		mcp.WithDescription("Searches for and retrieves relevant information from a knowledge base (memory RAG) based on a user's query. Use this tool to answer questions or find context on a specific topic by querying the stored information."),
		mcp.WithString("query", mcp.Required(), mcp.Description("The query to ask the memory RAG.")),
	)
	s.AddTool(queryTool, h.query)
}

func (h *handlers) memorize(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText("unimplemented"), nil
}

func (h *handlers) query(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText(sampleContext), nil
}
