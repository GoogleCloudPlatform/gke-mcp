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

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
	"github.com/GoogleCloudPlatform/gke-mcp/pkg/install"
	"github.com/GoogleCloudPlatform/gke-mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/server"
)

const (
	version = "0.0.1"
)

func main() {
	if len(os.Args) < 2 {
		startMCPServer()
		return
	}

	switch os.Args[1] {
	case "install":
		installCmd := flag.NewFlagSet("install", flag.ExitOnError)
		installCmd.Parse(os.Args[2:])

		if installCmd.NArg() != 1 {
			fmt.Fprintln(os.Stderr, "Error: `install` command requires exactly one argument: the tool name.")
			fmt.Fprintln(os.Stderr, "Usage: gke-mcp install <tool>")
			os.Exit(1)
		}

		wd, err := os.Getwd()
		if err != nil {
			log.Fatalf("Failed to get current working directory: %v", err)
		}

		exePath, err := os.Executable()
		if err != nil {
			log.Fatalf("Failed to get executable path: %v", err)
		}

		tool := installCmd.Arg(0)
		if tool != "gemini-cli" {
			fmt.Fprintf(os.Stderr, "Error: Unknown tool for install command: %s\n", tool)
			fmt.Fprintln(os.Stderr, "The only supported tool is 'gemini-cli'.")
			os.Exit(1)
		}

		if err := install.GeminiCLIExtension(wd, version, exePath); err != nil {
			log.Fatalf("Failed to install for gemini-cli: %v", err)
		}
		fmt.Println("Successfully installed GKE MCP server as a gemini-cli extension.")
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown command %q\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func startMCPServer() {
	s := server.NewMCPServer(
		"GKE MCP Server",
		version,
		server.WithToolCapabilities(true),
	)

	c := config.New(version)
	tools.Install(s, c)

	log.Printf("Starting GKE MCP Server")
	if err := server.ServeStdio(s); err != nil {
		log.Printf("Server error: %v\n", err)
	}
}

func printUsage() {
	fmt.Println("Usage: gke-mcp <command>")
	fmt.Println("")
	fmt.Println("Available commands:")
	fmt.Println("  install <tool>  Install a tool (currently only 'gemini-cli' is supported)")
	fmt.Println("  help            Show this help message")
	fmt.Println("")
	fmt.Println("If no command is specified, the GKE MCP server will be started.")
}
