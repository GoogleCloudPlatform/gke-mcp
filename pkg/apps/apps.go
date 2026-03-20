// Package apps registers MCP apps for GKE workflows.
package apps

import (
	"context"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/apps/dropdown"
	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type installer func(ctx context.Context, s *mcp.Server, c *config.Config) error

// InstallApps registers MCP tools that require a client host with 'apps' extension support.
func InstallApps(ctx context.Context, s *mcp.Server, c *config.Config) error {
	installers := []installer{
		dropdown.Install,
	}

	for _, installer := range installers {
		if err := installer(ctx, s, c); err != nil {
			return err
		}
	}

	return nil
}
