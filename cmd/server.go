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
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
	"github.com/GoogleCloudPlatform/gke-mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/server"
)

const (
	version = "0.0.1"
)

var (
	installGeminiFlag = flag.Bool("install_gemini", false, "Install this MCP Server into the Gemini settings file")
)

func main() {

	flag.Parse()

	if *installGeminiFlag {
		installGemini()
		return
	}

	// Create a new MCP server
	s := server.NewMCPServer(
		"GKE MCP Server",
		version,
		server.WithToolCapabilities(true),
	)

	c := config.New(version)
	tools.Install(s, c)

	// Start the stdio server
	log.Printf("Starting GKE MCP Server")
	if err := server.ServeStdio(s); err != nil {
		log.Printf("Server error: %v\n", err)
	}
}

func gkeMcpServer() map[string]interface{} {

	wd, _ := os.Getwd()

	return map[string]interface{}{
		"cwd":     wd,
		"command": "sh",
		"args": []string{
			"./run_mcp_server.sh",
		},
	}
}

func installGemini() {
	geminiSettingsFile := os.Getenv("HOME") + "/.gemini/settings.json"
	b, err := os.ReadFile(geminiSettingsFile)
	if err != nil {
		log.Printf("Failed to read Gemini settings file: %v", err)
		return
	}

	var settings map[string]interface{}
	err = json.Unmarshal([]byte(b), &settings)
	if err != nil {
		log.Printf("Failed to parse Gemini settings file: %v", err)
		return
	}

	if settings["mcpServers"] == nil {
		settings["mcpServers"] = map[string]interface{}{}
	}
	mcpServers, ok := settings["mcpServers"].(map[string]interface{})
	if !ok {
		log.Printf("Failed to parse mcpServers in Gemini settings file: %v", err)
		return
	}
	mcpServers["gke"] = gkeMcpServer()

	b, err = json.MarshalIndent(settings, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal Gemini settings file: %v", err)
		return
	}
	b = append(b, '\n')

	err = os.WriteFile(geminiSettingsFile, b, 0)
	if err != nil {
		log.Printf("Failed to write Gemini settings file: %v", err)
		return
	}
	log.Printf("Gemini settings updated! (%v)\n", geminiSettingsFile)
}
