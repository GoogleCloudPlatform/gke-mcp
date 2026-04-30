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

package cluster

import (
	"context"
	"fmt"

	containerpb "cloud.google.com/go/container/apiv1/containerpb"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/protobuf/encoding/protojson"
)

type listOperationsArgs struct {
	Parent string `json:"parent" jsonschema:"Required. The parent (project and location) where the operations will be listed. Specified in the format projects/*/locations/*."`
}

type getOperationArgs struct {
	Name string `json:"name" jsonschema:"Required. The name (project, location, operation id) of the operation to get. Specified in the format projects/*/locations/*/operations/*."`
}

type cancelOperationArgs struct {
	Name   string `json:"name" jsonschema:"Required. The name (project, location, operation id) of the operation to cancel. Specified in the format projects/*/locations/*/operations/*."`
	Parent string `json:"parent" jsonschema:"Required. The parent cluster of the operation. Specified in the format projects/*/locations/*/clusters/*."`
}

func (h *handlers) listOperations(ctx context.Context, _ *mcp.CallToolRequest, args *listOperationsArgs) (*mcp.CallToolResult, any, error) {
	if h.cmClient == nil {
		return nil, nil, fmt.Errorf("client not initialized")
	}
	req := &containerpb.ListOperationsRequest{
		Parent: args.Parent,
	}
	resp, err := h.cmClient.ListOperations(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: protojson.Format(resp)},
		},
	}, nil, nil
}

func (h *handlers) getOperation(ctx context.Context, _ *mcp.CallToolRequest, args *getOperationArgs) (*mcp.CallToolResult, any, error) {
	if h.cmClient == nil {
		return nil, nil, fmt.Errorf("client not initialized")
	}
	req := &containerpb.GetOperationRequest{
		Name: args.Name,
	}
	resp, err := h.cmClient.GetOperation(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: protojson.Format(resp)},
		},
	}, nil, nil
}

func (h *handlers) cancelOperation(ctx context.Context, _ *mcp.CallToolRequest, args *cancelOperationArgs) (*mcp.CallToolResult, any, error) {
	if h.cmClient == nil {
		return nil, nil, fmt.Errorf("client not initialized")
	}
	req := &containerpb.CancelOperationRequest{
		Name: args.Name,
	}
	err := h.cmClient.CancelOperation(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Operation %s cancelled successfully.", args.Name)},
		},
	}, nil, nil
}
