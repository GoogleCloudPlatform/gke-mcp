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

package dropdown

import (
	"context"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
	"github.com/GoogleCloudPlatform/gke-mcp/ui"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	resourceURI = "ui://dropdown/index.html"
	mimeType    = "text/html;profile=mcp-app"

	StatusPendingUserInput = "PENDING_USER_INPUT"
)

type dropdownArgs struct {
	Title   string   `json:"title,omitempty" jsonschema:"Title to display above the dropdown"`
	Options []string `json:"options" jsonschema:"List of resources to display in the dropdown"`
}

type PendingResponse struct {
	Status  string   `json:"status"`
	Options []string `json:"options"`
	Message string   `json:"message"`
}

// Install registers the dropdown tool with the MCP server.
func Install(ctx context.Context, s *mcp.Server, c *config.Config) error {
	mcp.AddTool(s, &mcp.Tool{
		Name: "dropdown",
		Description: `Renders an interactive UI dropdown for the user to select an item from a list.
Use this tool when you need the user to choose one option from a set of available resources (e.g., clusters, regions, namespaces).
You MUST provide a valid array of 1 or more options. 

Timing: Call this tool immediately before you need the user's input to proceed. Do not ask the user for clarification in plain text; calling this tool serves as your question to the user.
After calling this tool, STOP and wait for the user to make a selection via the UI.
Do NOT list the options in your text response; the UI itself serves as the list and confirmation.`,
		Meta: mcp.Meta{
			"ui": map[string]interface{}{
				"resourceUri": resourceURI,
				"visibility":  []string{"app"},
			},
		},
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"title": map[string]interface{}{
					"type":        "string",
					"description": "Title to display above the dropdown",
				},
				"options": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "string",
					},
					"description": "List of resources to display in the dropdown",
				},
			},
			"required": []string{"options"},
		},
	}, dropdownHandler)

	s.AddResource(&mcp.Resource{
		Name:     "GKE Resource Dropdown UI",
		URI:      resourceURI,
		MIMEType: mimeType,
	}, func(ctx context.Context, request *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{
					URI:      resourceURI,
					MIMEType: mimeType,
					Text:     string(ui.DropdownHTML),
				},
			},
		}, nil
	})

	return nil
}

func dropdownHandler(ctx context.Context, request *mcp.CallToolRequest, args *dropdownArgs) (*mcp.CallToolResult, any, error) {
	payload := PendingResponse{
		Status:  StatusPendingUserInput,
		Options: args.Options,
		Message: "Present these options to the user. Wait until selection is made",
	}

	return &mcp.CallToolResult{
		StructuredContent: payload,
	}, nil, nil
}
