// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package install

import (
	_ "embed"
	"fmt"
	"os"
)

type InstallOptions struct {
	version       string
	installDir    string
	exePath       string
	developerMode bool
}

func NewInstallOptions(
	version string,
	projectOnly bool,
	developerMode bool,
) (*InstallOptions, error) {

	installDir := ""
	var err error
	if projectOnly {
		installDir, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current working directory: %w", err)
		}
	} else {
		installDir, err = os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("could not determine home directory: %w", err)
		}
	}

	exePath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable path: %w", err)
	}

	return &InstallOptions{
		version:       version,
		installDir:    installDir,
		exePath:       exePath,
		developerMode: developerMode,
	}, nil
}

//go:embed GEMINI.md
var GeminiMarkdown []byte
