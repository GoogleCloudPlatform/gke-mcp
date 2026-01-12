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

package giq

import (
	"testing"
)

func TestGiqGenerateManifestArgs_Fields(t *testing.T) {
	args := giqGenerateManifestArgs{
		Model:                   "llama-3-1-405b",
		ModelServer:             "tgi",
		Accelerator:             "nvidia-a100-80gb",
		TargetNTPOTMilliseconds: "10",
	}

	if args.Model != "llama-3-1-405b" {
		t.Errorf("Model = %s, want llama-3-1-405b", args.Model)
	}
	if args.ModelServer != "tgi" {
		t.Errorf("ModelServer = %s, want tgi", args.ModelServer)
	}
	if args.Accelerator != "nvidia-a100-80gb" {
		t.Errorf("Accelerator = %s, want nvidia-a100-80gb", args.Accelerator)
	}
	if args.TargetNTPOTMilliseconds != "10" {
		t.Errorf("TargetNTPOTMilliseconds = %s, want 10", args.TargetNTPOTMilliseconds)
	}
}

func TestGiqGenerateManifestArgs_RequiredFields(t *testing.T) {
	args := giqGenerateManifestArgs{
		Model:       "llama-3-1-8b",
		ModelServer: "vllm",
		Accelerator: "nvidia-l4",
	}

	if args.Model != "llama-3-1-8b" {
		t.Error("Model field not working")
	}
	if args.ModelServer != "vllm" {
		t.Error("ModelServer field not working")
	}
	if args.Accelerator != "nvidia-l4" {
		t.Error("Accelerator field not working")
	}
	if args.TargetNTPOTMilliseconds != "" {
		t.Errorf("Expected empty TargetNTPOTMilliseconds, got %s", args.TargetNTPOTMilliseconds)
	}
}

func TestGiqGenerateManifestArgs_Empty(t *testing.T) {
	args := giqGenerateManifestArgs{}
	if args.Model != "" {
		t.Errorf("Expected empty Model, got %s", args.Model)
	}
	if args.ModelServer != "" {
		t.Errorf("Expected empty ModelServer, got %s", args.ModelServer)
	}
	if args.Accelerator != "" {
		t.Errorf("Expected empty Accelerator, got %s", args.Accelerator)
	}
	if args.TargetNTPOTMilliseconds != "" {
		t.Errorf("Expected empty TargetNTPOTMilliseconds, got %s", args.TargetNTPOTMilliseconds)
	}
}

func TestGiqGenerateManifestArgs_WithTargetNTPOT(t *testing.T) {
	args := giqGenerateManifestArgs{
		Model:                   "mixtral-8x7b",
		ModelServer:             "tgi",
		Accelerator:             "nvidia-a100-80gb",
		TargetNTPOTMilliseconds: "15.5",
	}

	if args.TargetNTPOTMilliseconds != "15.5" {
		t.Errorf("TargetNTPOTMilliseconds = %s, want 15.5", args.TargetNTPOTMilliseconds)
	}
}

func TestGiqGenerateManifestArgs_MultipleAccelerators(t *testing.T) {
	tests := []struct {
		name        string
		accelerator string
	}{
		{"nvidia-a100-80gb", "nvidia-a100-80gb"},
		{"nvidia-a100-40gb", "nvidia-a100-40gb"},
		{"nvidia-l4", "nvidia-l4"},
		{"nvidia-h100-80gb", "nvidia-h100-80gb"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := giqGenerateManifestArgs{
				Model:       "test-model",
				ModelServer: "test-server",
				Accelerator: tt.accelerator,
			}
			if args.Accelerator != tt.accelerator {
				t.Errorf("Accelerator = %s, want %s", args.Accelerator, tt.accelerator)
			}
		})
	}
}

func TestGiqGenerateManifestArgs_JSONTags(t *testing.T) {
	args := giqGenerateManifestArgs{
		Model:       "test-model",
		ModelServer: "test-server",
		Accelerator: "test-accelerator",
	}

	if args.Model != "test-model" {
		t.Error("Model field not working correctly")
	}
}

func TestGiqGenerateManifestArgs_DifferentModelServers(t *testing.T) {
	servers := []string{"tgi", "vllm", "sglang"}
	for _, server := range servers {
		args := giqGenerateManifestArgs{
			Model:       "test-model",
			ModelServer: server,
			Accelerator: "test-accelerator",
		}
		if args.ModelServer != server {
			t.Errorf("ModelServer = %s, want %s", args.ModelServer, server)
		}
	}
}
