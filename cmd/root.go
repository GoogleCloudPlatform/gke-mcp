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
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
	"google.golang.org/api/option"
)

var (
	version = "(unknown)"

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

	installDeveloper bool
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

	rootCmd.AddCommand(installCmd)
	installCmd.AddCommand(installGeminiCLICmd)
	installCmd.PersistentFlags().BoolVarP(&installDeveloper, "developer", "d", false, "Install the MCP Server in developer mode")
}

func runRootCmd(cmd *cobra.Command, args []string) {
	startMCPServer(cmd.Context())
}

func startMCPServer(ctx context.Context) {
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
		server.WithInstructions(instructions),
	)

	if err := tools.Install(ctx, s, c); err != nil {
		log.Fatalf("Failed to install tools: %v\n", err)
	}

	log.Printf("Starting GKE MCP Server (%s)", version)
	if err := server.ServeStdio(s); err != nil {
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

func runInstallGeminiCLICmd(cmd *cobra.Command, args []string) {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current working directory: %v", err)
	}

	exePath, err := os.Executable()
	if err != nil {
		log.Fatalf("Failed to get executable path: %v", err)
	}

	if err := install.GeminiCLIExtension(wd, version, exePath, installDeveloper); err != nil {
		log.Fatalf("Failed to install for gemini-cli: %v", err)
	}
	fmt.Println("Successfully installed GKE MCP server as a gemini-cli extension.")
}
