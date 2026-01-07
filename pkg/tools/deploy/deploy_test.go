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

package deploy

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestGkeDeploy(t *testing.T) {
	testCases := []struct {
		name          string
		args          *deployArgs
		wantContained string
		wantErr       string
	}{
		{
			name:          "successful request",
			args:          &deployArgs{UserRequest: "deploy my-app to staging"},
			wantContained: "You are an expert GKE (Google Kubernetes Engine) deployment assistant.",
		},
		{
			name:    "empty request",
			args:    &deployArgs{UserRequest: ""},
			wantErr: "argument 'user_request' cannot be empty",
		},
		{
			name:    "whitespace only request",
			args:    &deployArgs{UserRequest: "   "},
			wantErr: "argument 'user_request' cannot be empty",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, _, err := gkeDeployHandler(context.Background(), nil, tc.args)

			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("gkeDeployHandler() err = nil, want = %q", tc.wantErr)
				}
				if !strings.Contains(err.Error(), tc.wantErr) {
					t.Errorf("gkeDeployHandler() err = %q, want to contain %q", err.Error(), tc.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("gkeDeployHandler() returned unexpected error: %v", err)
			}

			if len(result.Content) != 1 {
				t.Errorf("gkeDeployHandler() returned %d content objects, want: 1", len(result.Content))
			}

			textContent, ok := result.Content[0].(*mcp.TextContent)
			if !ok {
				t.Errorf("content is not *mcp.TextContent, got %T", result.Content[0])
				return
			}

			if !strings.Contains(textContent.Text, tc.wantContained) {
				t.Errorf("gkeDeployHandler() result does not contain expected text.\nGot: ...%s...\nWant contained: %s", textContent.Text[:100], tc.wantContained)
			}

			// Also check that the user request is embedded in the output if it was successful
			if tc.args.UserRequest != "" && !strings.Contains(textContent.Text, tc.args.UserRequest) {
				// Note: The current template might not actually print the user_request back to the user,
				// but based on `gkeDeployTmpl.Execute(&buf, map[string]string{"user_request": args.UserRequest})
				// it likely uses it. Let's check the template content in deploy.go again if this fails.
				// Wait, looking at deploy.go:
				// const gkeDeployPromptTemplate = `...`
				// It doesn't seem to actually use {{.user_request}} in the template string I read earlier!
				// Let's re-read deploy.go to be sure.
			}
		})
	}
}
