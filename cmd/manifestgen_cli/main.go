package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/agents/manifestgen"
	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <prompt>", os.Args[0])
	}
	prompt := os.Args[1]

	ctx := context.Background()
	// Using a dummy version string for the config.
	cfg := config.New("0.0.1") 

	agent, err := manifestgen.NewAgent(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	manifest, err := agent.GenerateManifest(ctx, prompt)
	if err != nil {
		log.Fatalf("Failed to generate manifest: %v", err)
	}

	fmt.Println(manifest)
}
