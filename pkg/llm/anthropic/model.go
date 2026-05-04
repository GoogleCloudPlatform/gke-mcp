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

package anthropic

import (
	"context"
	"fmt"
	"iter"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"google.golang.org/adk/model"
	"google.golang.org/genai"
)

// AnthropicModel implements the model.LLM interface for Anthropic Claude models.
type AnthropicModel struct {
	client anthropic.Client
	model  string
}

// NewModel creates a new AnthropicModel instance.
func NewModel(apiKey string, modelName string) *AnthropicModel {
	return &AnthropicModel{
		client: anthropic.NewClient(option.WithAPIKey(apiKey)),
		model:  modelName,
	}
}

// Name returns the name of the provider.
func (m *AnthropicModel) Name() string {
	return "anthropic"
}

// GenerateContent implements the model.LLM interface to generate content.
func (m *AnthropicModel) GenerateContent(ctx context.Context, req *model.LLMRequest, stream bool) iter.Seq2[*model.LLMResponse, error] {
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
			MaxTokens: 1024, // Fallback default
			Messages:  messages,
		}
		if systemPrompt != "" {
			params.System = []anthropic.TextBlockParam{{Text: systemPrompt}}
		}
		if req.Config != nil && req.Config.MaxOutputTokens > 0 {
			params.MaxTokens = int64(req.Config.MaxOutputTokens)
		}

		// 4. Call API
		if stream {
			// TODO: Implement streaming
			yield(nil, fmt.Errorf("streaming not implemented yet for Anthropic adapter"))
			return
		}

		resp, err := m.client.Messages.New(ctx, params)
		if err != nil {
			yield(nil, fmt.Errorf("anthropic api error: %w", err))
			return
		}

		// 5. Map response back
		var responseText string
		for _, block := range resp.Content {
			if block.Type == "text" {
				responseText += block.Text
			}
		}

		adkResp := &model.LLMResponse{
			Content: &genai.Content{
				Role: "model",
				Parts: []*genai.Part{
					{
						Text: responseText,
					},
				},
			},
		}

		yield(adkResp, nil)
	}
}

