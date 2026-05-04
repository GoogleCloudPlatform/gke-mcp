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

// Package giq provides tools for GKE Inference Quickstart workflows.
package giq

import (
	"context"
	"fmt"
	"strings"

	gkerecommender "cloud.google.com/go/gkerecommender/apiv1"
	gkerecommenderpb "cloud.google.com/go/gkerecommender/apiv1/gkerecommenderpb"
	"google.golang.org/api/iterator"
)

// GenerateInferenceManifestArgs holds arguments for generating a GKE Inference Quickstart manifest.
type GenerateInferenceManifestArgs struct {
	Model                   string `json:"model" jsonschema:"The model to use. Get the list of valid models from 'gcloud container ai profiles models list' if the user doesn't provide it."`
	ModelServer             string `json:"model_server" jsonschema:"The model server to use. Get the list of valid model servers from 'gcloud container ai profiles list --format='table(modelServerInfo.model,modelServerInfo.modelServer,modelServerInfo.modelServerVersion,acceleratorType)' if the user doesn't provide it. You can filter that gcloud command on '--model={model}' if the user provides the model."`
	Accelerator             string `json:"accelerator" jsonschema:"The accelerator to use. Get the list of valid accelerators from 'gcloud container ai profiles list --format='table(modelServerInfo.model,modelServerInfo.modelServer,modelServerInfo.modelServerVersion,acceleratorType)' if the user doesn't provide it. You can filter that gcloud command on '--model={model}' and '--model-server={model-server}' if the user provides those values."`
	TargetNTPOTMilliseconds string `json:"target_ntpot_milliseconds,omitempty" jsonschema:"The maximum normalized time per output token (NTPOT) in milliseconds.NTPOT is measured as the request_latency / output_tokens."`
}

// GenerateInferenceManifest core logic for GKE Inference Quickstart manifest generation.
func GenerateInferenceManifest(ctx context.Context, args *GenerateInferenceManifestArgs) (string, error) {
	if args == nil {
		return "", fmt.Errorf("args cannot be nil")
	}
	if args.Model == "" {
		return "", fmt.Errorf("model argument cannot be empty")
	}
	if args.ModelServer == "" {
		return "", fmt.Errorf("model_server argument cannot be empty")
	}
	if args.Accelerator == "" {
		return "", fmt.Errorf("accelerator argument cannot be empty")
	}

	client, err := gkerecommender.NewGkeInferenceQuickstartClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create gkerecommender client: %w", err)
	}
	defer func() {
		_ = client.Close()
	}()

	req := &gkerecommenderpb.GenerateOptimizedManifestRequest{
		ModelServerInfo: &gkerecommenderpb.ModelServerInfo{
			Model:       args.Model,
			ModelServer: args.ModelServer,
		},
		AcceleratorType: args.Accelerator,
	}

	resp, err := client.GenerateOptimizedManifest(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to generate optimized manifest via SDK: %w", err)
	}

	var manifests []string
	for _, m := range resp.GetKubernetesManifests() {
		manifests = append(manifests, m.GetContent())
	}

	return strings.Join(manifests, "\n---\n"), nil
}

var fetchModelsFunc = func(ctx context.Context) ([]string, error) {
	client, err := gkerecommender.NewGkeInferenceQuickstartClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create gkerecommender client: %w", err)
	}
	defer func() {
		_ = client.Close()
	}()

	req := &gkerecommenderpb.FetchModelsRequest{}
	it := client.FetchModels(ctx, req)

	var models []string
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to fetch next model: %w", err)
		}
		models = append(models, resp)
	}
	return models, nil
}

// FetchModels fetches available models for GKE.
func FetchModels(ctx context.Context) (string, error) {
	// TODO: Add pagination support once model list becomes very large to avoid memory risks.
	models, err := fetchModelsFunc(ctx)
	if err != nil {
		return "", err
	}
	return strings.Join(models, "\n"), nil
}

var fetchModelServersFunc = func(ctx context.Context, model string) ([]string, error) {
	client, err := gkerecommender.NewGkeInferenceQuickstartClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create gkerecommender client: %w", err)
	}
	defer func() {
		_ = client.Close()
	}()

	req := &gkerecommenderpb.FetchModelServersRequest{
		Model: model,
	}
	it := client.FetchModelServers(ctx, req)

	var servers []string
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to fetch next model server: %w", err)
		}
		servers = append(servers, resp)
	}
	return servers, nil
}

// FetchModelServers fetches available model servers for a given model.
func FetchModelServers(ctx context.Context, model string) (string, error) {
	if model == "" {
		return "", fmt.Errorf("model argument cannot be empty")
	}
	servers, err := fetchModelServersFunc(ctx, model)
	if err != nil {
		return "", err
	}
	return strings.Join(servers, "\n"), nil
}

var fetchProfilesFunc = func(ctx context.Context, model, modelServer, modelServerVersion string) ([]*gkerecommenderpb.Profile, error) {
	client, err := gkerecommender.NewGkeInferenceQuickstartClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create gkerecommender client: %w", err)
	}
	defer func() {
		_ = client.Close()
	}()

	req := &gkerecommenderpb.FetchProfilesRequest{
		Model:              model,
		ModelServer:        modelServer,
		ModelServerVersion: modelServerVersion,
	}
	it := client.FetchProfiles(ctx, req)

	var profiles []*gkerecommenderpb.Profile
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to fetch next profile: %w", err)
		}
		profiles = append(profiles, resp)
	}
	return profiles, nil
}

// FetchProfiles fetches available profiles for GKE.
func FetchProfiles(ctx context.Context, model, modelServer, modelServerVersion string) (string, error) {
	// TODO: Add pagination support once profile list becomes very large to avoid memory risks.
	profiles, err := fetchProfilesFunc(ctx, model, modelServer, modelServerVersion)
	if err != nil {
		return "", err
	}
	var output []string
	for _, p := range profiles {
		output = append(output, p.String())
	}
	return strings.Join(output, "\n---\n"), nil
}

var fetchModelServerVersionsFunc = func(ctx context.Context, model, modelServer string) ([]string, error) {
	client, err := gkerecommender.NewGkeInferenceQuickstartClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create gkerecommender client: %w", err)
	}
	defer func() {
		_ = client.Close()
	}()

	req := &gkerecommenderpb.FetchModelServerVersionsRequest{
		Model:       model,
		ModelServer: modelServer,
	}
	it := client.FetchModelServerVersions(ctx, req)

	var versions []string
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to fetch next model server version: %w", err)
		}
		versions = append(versions, resp)
	}
	return versions, nil
}

// FetchModelServerVersions fetches available model server versions for a given model and model server.
func FetchModelServerVersions(ctx context.Context, model, modelServer string) (string, error) {
	if model == "" {
		return "", fmt.Errorf("model argument cannot be empty")
	}
	if modelServer == "" {
		return "", fmt.Errorf("model_server argument cannot be empty")
	}
	versions, err := fetchModelServerVersionsFunc(ctx, model, modelServer)
	if err != nil {
		return "", err
	}
	return strings.Join(versions, "\n"), nil
}
