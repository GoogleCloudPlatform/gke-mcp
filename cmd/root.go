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

package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"strings"

	container "cloud.google.com/go/container/apiv1"
	"cloud.google.com/go/container/apiv1/containerpb"
	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
	"github.com/GoogleCloudPlatform/gke-mcp/pkg/install"
	"github.com/GoogleCloudPlatform/gke-mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
	"google.golang.org/api/option"
)

const (
	geminiInstructionsURI = "mcp://gke/pkg/install/GEMINI.md"
)

var (
	version = "(unknown)"

	// command flags
	serverMode string
	serverPort int

	// rootCmd represents the base command when called without any subcommands
	rootCmd = &cobra.Command{
		Use:   "gke-mcp",
		Short: "An MCP Server for Google Kubernetes Engine",
		Run:   runRootCmd,
	}

	installCmd = &cobra.Command{
		Use:   "install",
		Short: "Install the GKE MCP Server into your AI tool settings.",
	}

	installGeminiCLICmd = &cobra.Command{
		Use:   "gemini-cli",
		Short: "Install the GKE MCP Server into your Gemini CLI settings.",
		Run:   runInstallGeminiCLICmd,
	}

	installCursorCmd = &cobra.Command{
		Use:   "cursor",
		Short: "Install the GKE MCP Server into your Cursor settings.",
		Run:   runInstallCursorCmd,
	}

	installClaudeDesktopCmd = &cobra.Command{
		Use:   "claude-desktop",
		Short: "Install the GKE MCP Server into your Claude Desktop settings.",
		Run:   runInstallClaudeDesktopCmd,
	}

	installDeveloper   bool
	installProjectOnly bool
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	if bi, ok := debug.ReadBuildInfo(); ok {
		version = bi.Main.Version
	} else {
		log.Printf("Failed to read build info to get version.")
	}

	rootCmd.Flags().StringVar(&serverMode, "server-mode", "stdio", "transport to use for the server: stdio (default) or http")
	rootCmd.Flags().IntVar(&serverPort, "server-port", 8080, "server port to use when server-mode is http; defaults to 8080")
	rootCmd.AddCommand(installCmd)

	installCmd.AddCommand(installGeminiCLICmd)
	installCmd.AddCommand(installCursorCmd)
	installCmd.AddCommand(installClaudeDesktopCmd)

	installGeminiCLICmd.Flags().BoolVarP(&installDeveloper, "developer", "d", false, "Install the MCP Server in developer mode for Gemini CLI")
	installGeminiCLICmd.Flags().BoolVarP(&installProjectOnly, "project-only", "p", false, "Install the MCP Server only for the current project. Please run this in the root directory of your project")

	installCursorCmd.Flags().BoolVarP(&installProjectOnly, "project-only", "p", false, "Install the MCP Server only for the current project. Please run this in the root directory of your project")
}

type startOptions struct {
	serverMode string
	serverPort int
}

func runRootCmd(cmd *cobra.Command, args []string) {
	opts := startOptions{
		serverMode: serverMode,
		serverPort: serverPort,
	}
	startMCPServer(cmd.Context(), opts)
}

func startMCPServer(ctx context.Context, opts startOptions) {
	c := config.New(version)

	instructions := ""
	if err := adcAuthCheck(ctx, c); err != nil {
		if strings.Contains(err.Error(), "Unauthenticated") {
			log.Printf("GKE API calls requires Application Default Credentials (https://cloud.google.com/docs/authentication/application-default-credentials). Get credentials with `gcloud auth application-default login` before calling MCP tools.")
			instructions += "GKE API calls requires Application Default Credentials (https://cloud.google.com/docs/authentication/application-default-credentials). Get credentials with `gcloud auth application-default login` before calling MCP tools."
		}
	}

	s := server.NewMCPServer(
		"GKE MCP Server",
		version,
		server.WithToolCapabilities(true),
		server.WithResourceCapabilities(false, false),
		server.WithInstructions(instructions),
	)

	resource := mcp.NewResource(
		geminiInstructionsURI,
		"GEMINI.md",
		mcp.WithResourceDescription("Instructions for how to use the GKE MCP server"),
		mcp.WithMIMEType("text/markdown"),
	)

	s.AddResource(resource, func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      geminiInstructionsURI,
				MIMEType: "text/markdown",
				Text:     string(install.GeminiMarkdown),
			},
		}, nil
	})

	s.AddPrompt(GkeCostPrompt, GkeCostHandler)

	if err := tools.Install(ctx, s, c); err != nil {
		log.Fatalf("Failed to install tools: %v\n", err)
	}

	// start server in the right mode
	log.Printf("Starting GKE MCP Server (%s) in mode '%s'", version, opts.serverMode)
	var err error
	endpoint := fmt.Sprintf(":%d", opts.serverPort)

	switch opts.serverMode {
	case "stdio":
		err = server.ServeStdio(s)
	case "http":
		httpServer := server.NewStreamableHTTPServer(s)
		log.Printf("Listening for HTTP connections on port: %d", opts.serverPort)
		err = httpServer.Start(endpoint)
	default:
		log.Printf("Unknown mode '%s', defaulting to 'stdio'", opts.serverMode)
		err = server.ServeStdio(s)
	}
	if err != nil {
		if errors.Is(err, context.Canceled) {
			log.Printf("Server shutting down.")
		} else {
			log.Printf("Server error: %v\n", err)
		}
	}
}

func adcAuthCheck(ctx context.Context, c *config.Config) error {
	projectID := c.DefaultProjectID()
	// Can't do a pre-flight check without a default project.
	if projectID == "" {
		return nil
	}

	location := c.DefaultLocation()
	// Without a default location try checking us-central1.
	if location == "" {
		location = "us-central1"
	}

	cmClient, err := container.NewClusterManagerClient(ctx, option.WithUserAgent(c.UserAgent()))
	if err != nil {
		return fmt.Errorf("failed to create cluster manager client: %w", err)
	}
	defer cmClient.Close()

	_, err = cmClient.GetServerConfig(ctx, &containerpb.GetServerConfigRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s", projectID, location),
	})
	return err
}

func installOptions() (*install.InstallOptions, error) {
	return install.NewInstallOptions(
		version,
		installProjectOnly,
		installDeveloper,
	)
}

func runInstallGeminiCLICmd(cmd *cobra.Command, args []string) {
	opts, err := installOptions()
	if err != nil {
		log.Fatalf("Failed to get install options: %v", err)
	}

	if err := install.GeminiCLIExtension(opts); err != nil {
		log.Fatalf("Failed to install for gemini-cli: %v", err)
	}
	fmt.Println("Successfully installed GKE MCP server as a gemini-cli extension.")
}

func runInstallCursorCmd(cmd *cobra.Command, args []string) {
	opts, err := installOptions()
	if err != nil {
		log.Fatalf("Failed to get install options: %v", err)
	}

	if err := install.CursorMCPExtension(opts); err != nil {
		log.Fatalf("Failed to install for cursor: %v", err)
	}
	fmt.Println("Successfully installed GKE MCP server as a cursor MCP server.")
}

func runInstallClaudeDesktopCmd(cmd *cobra.Command, args []string) {
	opts, err := installOptions()
	if err != nil {
		log.Fatalf("Failed to get install options: %v", err)
	}

	if err := install.ClaudeDesktopExtension(opts); err != nil {
		log.Fatalf("Failed to install for Claude Desktop: %v", err)
	}
	fmt.Println("Successfully installed GKE MCP server in Claude Desktop configuration.")
}
