// Copyright 2026 Google LLC
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

// Package anthropic provides an ADK model adapter for Anthropic Claude models.
package anthropic

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"google.golang.org/adk/model"
	"google.golang.org/genai"
)

const defaultMaxTokens = 1024

// Model implements the model.LLM interface for Anthropic Claude models.
type Model struct {
	client anthropic.Client
	model  string
}

// NewModel creates a new Model instance.
func NewModel(apiKey string, modelName string) *Model {
	return &Model{
		client: anthropic.NewClient(option.WithAPIKey(apiKey)),
		model:  modelName,
	}
}

// Name returns the name of the provider.
func (m *Model) Name() string {
	return "anthropic"
}

// GenerateContent implements the model.LLM interface to generate content.
func (m *Model) GenerateContent(ctx context.Context, req *model.LLMRequest, stream bool) iter.Seq2[*model.LLMResponse, error] {
	return func(yield func(*model.LLMResponse, error) bool) {
		// 1. Map System Instruction
		var systemPrompt string
		if req.Config != nil && req.Config.SystemInstruction != nil {
			for _, part := range req.Config.SystemInstruction.Parts {
				if part.Text != "" {
					systemPrompt += part.Text
				}
			}
		}

		// 2. Map Messages
		var messages []anthropic.MessageParam
		for _, content := range req.Contents {
			var role anthropic.MessageParamRole
			switch content.Role {
			case "user":
				role = anthropic.MessageParamRoleUser
			case "model":
				role = anthropic.MessageParamRoleAssistant
			default:
				role = anthropic.MessageParamRoleUser
			}

			var blocks []anthropic.ContentBlockParamUnion
			for _, part := range content.Parts {
				if part.Text != "" {
					blocks = append(blocks, anthropic.NewTextBlock(part.Text))
				} else if part.FunctionResponse != nil {
					// Map ADK FunctionResponse to Anthropic tool_result block
					contentStr := ""
					if part.FunctionResponse.Response != nil {
						b, _ := json.Marshal(part.FunctionResponse.Response)
						contentStr = string(b)
					}
					// Using NewToolResultBlock from Anthropic SDK
					blocks = append(blocks, anthropic.NewToolResultBlock(part.FunctionResponse.ID, contentStr, false))
				}
			}

			if len(blocks) > 0 {
				messages = append(messages, anthropic.MessageParam{
					Role:    role,
					Content: blocks,
				})
			}
		}

		// 3. Create request params
		params := anthropic.MessageNewParams{
			Model:     anthropic.Model(m.model),
			MaxTokens: defaultMaxTokens,
			Messages:  messages,
		}
		if systemPrompt != "" {
			params.System = []anthropic.TextBlockParam{{Text: systemPrompt}}
		}
		if req.Config != nil && req.Config.MaxOutputTokens > 0 {
			params.MaxTokens = int64(req.Config.MaxOutputTokens)
		}

		// 4. Map Tools
		if req.Config != nil && len(req.Config.Tools) > 0 {
			var tools []anthropic.ToolUnionParam
			for _, t := range req.Config.Tools {
				if t == nil {
					continue
				}
				for _, fd := range t.FunctionDeclarations {
					if fd == nil {
						continue
					}
					properties := map[string]any{}
					var required []string
					if fd.Parameters != nil {
						properties = schemaPropertiesToMap(fd.Parameters.Properties)
						required = fd.Parameters.Required
					}
					tools = append(tools, anthropic.ToolUnionParam{
						OfTool: &anthropic.ToolParam{
							Name:        fd.Name,
							Description: anthropic.String(fd.Description),
							InputSchema: anthropic.ToolInputSchemaParam{
								Properties: properties,
								Required:   required,
							},
						},
					})
				}
			}
			params.Tools = tools
		}

		// 5. Call API
		if stream {
			yield(nil, fmt.Errorf("streaming not implemented yet for Anthropic adapter"))
			return
		}

		resp, err := m.client.Messages.New(ctx, params)
		if err != nil {
			yield(nil, fmt.Errorf("anthropic api error: %w", err))
			return
		}

		// 6. Map response back
		var parts []*genai.Part
		for _, block := range resp.Content {
			if block.Type == "text" {
				parts = append(parts, &genai.Part{Text: block.Text})
			} else if block.Type == "tool_use" {
				args := make(map[string]any)
				if block.Input != nil {
					if err := json.Unmarshal(block.Input, &args); err != nil {
						yield(nil, fmt.Errorf("failed to unmarshal tool input: %w", err))
						return
					}
				}
				parts = append(parts, &genai.Part{
					FunctionCall: &genai.FunctionCall{
						ID:   block.ID,
						Name: block.Name,
						Args: args,
					},
				})
			}
		}

		adkResp := &model.LLMResponse{
			Content: &genai.Content{
				Role:  "model",
				Parts: parts,
			},
		}

		yield(adkResp, nil)
	}
}

// Helper function to convert schema properties
func schemaPropertiesToMap(props map[string]*genai.Schema) map[string]any {
	if props == nil {
		return nil
	}
	result := make(map[string]any)
	for name, schema := range props {
		if schema == nil {
			continue
		}
		result[name] = schemaToMap(schema)
	}
	return result
}

// Helper function to convert schema to map
func schemaToMap(schema *genai.Schema) map[string]any {
	if schema == nil {
		return nil
	}
	result := make(map[string]any)
	if schema.Type != "" {
		result["type"] = strings.ToLower(string(schema.Type))
	}
	if schema.Description != "" {
		result["description"] = schema.Description
	}
	if len(schema.Enum) > 0 {
		result["enum"] = schema.Enum
	}
	if schema.Items != nil {
		result["items"] = schemaToMap(schema.Items)
	}
	if len(schema.Properties) > 0 {
		result["properties"] = schemaPropertiesToMap(schema.Properties)
	}
	if len(schema.Required) > 0 {
		result["required"] = schema.Required
	}
	return result
}
