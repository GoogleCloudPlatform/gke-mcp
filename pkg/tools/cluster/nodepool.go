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
	"github.com/GoogleCloudPlatform/gke-mcp/pkg/tools/params"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/protobuf/encoding/protojson"
)

type createNodePoolArgs struct {
	params.Cluster
	NodePool string `json:"nodePool" jsonschema:"Required. The node pool to create represented as a string using JSON format."`
}

type listNodePoolsArgs struct {
	params.Cluster
}

type getNodePoolArgs struct {
	params.NodePool
}

type updateNodePoolArgs struct {
	params.NodePool
	Update string `json:"update" jsonschema:"Required. A node pool update request represented as a string using JSON format."`
}

type deleteNodePoolArgs struct {
	params.NodePool
}

func (h *handlers) createNodePool(ctx context.Context, _ *mcp.CallToolRequest, args *createNodePoolArgs) (*mcp.CallToolResult, any, error) {
	var nodePoolObj containerpb.NodePool
	if err := protojson.Unmarshal([]byte(args.NodePool), &nodePoolObj); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal node pool JSON: %w", err)
	}

	req := &containerpb.CreateNodePoolRequest{
		Parent:   args.ClusterPath(),
		NodePool: &nodePoolObj,
	}
	resp, err := h.cmClient.CreateNodePool(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: protojson.Format(resp)},
		},
	}, nil, nil
}

func (h *handlers) listNodePools(ctx context.Context, _ *mcp.CallToolRequest, args *listNodePoolsArgs) (*mcp.CallToolResult, any, error) {
	req := &containerpb.ListNodePoolsRequest{
		Parent: args.ClusterPath(),
	}
	resp, err := h.cmClient.ListNodePools(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: protojson.Format(resp)},
		},
	}, nil, nil
}

func (h *handlers) getNodePool(ctx context.Context, _ *mcp.CallToolRequest, args *getNodePoolArgs) (*mcp.CallToolResult, any, error) {
	req := &containerpb.GetNodePoolRequest{
		Name: args.NodePoolPath(),
	}
	resp, err := h.cmClient.GetNodePool(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: protojson.Format(resp)},
		},
	}, nil, nil
}

func (h *handlers) updateNodePool(ctx context.Context, _ *mcp.CallToolRequest, args *updateNodePoolArgs) (*mcp.CallToolResult, any, error) {
	var req containerpb.UpdateNodePoolRequest
	if err := protojson.Unmarshal([]byte(args.Update), &req); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal update JSON: %w", err)
	}
	req.Name = args.NodePoolPath()

	resp, err := h.cmClient.UpdateNodePool(ctx, &req)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: protojson.Format(resp)},
		},
	}, nil, nil
}

func (h *handlers) deleteNodePool(ctx context.Context, _ *mcp.CallToolRequest, args *deleteNodePoolArgs) (*mcp.CallToolResult, any, error) {
	req := &containerpb.DeleteNodePoolRequest{
		Name: args.NodePoolPath(),
	}
	resp, err := h.cmClient.DeleteNodePool(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: protojson.Format(resp)},
		},
	}, nil, nil
}
