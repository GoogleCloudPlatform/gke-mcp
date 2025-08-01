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

package config

import (
	"log"
	"os/exec"
	"strings"
)

type Config struct {
	userAgent        string
	defaultProjectID string
	defaultLocation  string
}

func (c *Config) UserAgent() string {
	return c.userAgent
}

func (c *Config) DefaultProjectID() string {
	return c.defaultProjectID
}

func (c *Config) DefaultLocation() string {
	return c.defaultLocation
}

func New(version string) *Config {
	return &Config{
		userAgent:        "gke-mcp/" + version,
		defaultProjectID: getDefaultProjectID(),
		defaultLocation:  getDefaultLocation(),
	}
}

func getDefaultProjectID() string {
	projectID, err := getGcloudConfig("core/project")
	if err != nil {
		log.Printf("Failed to get default project: %v", err)
		return ""
	}
	return projectID
}

func getDefaultLocation() string {
	region, err := getGcloudConfig("compute/region")
	if err == nil {
		return region
	}
	zone, err := getGcloudConfig("compute/zone")
	if err == nil {
		return zone
	}
	return ""
}

func getGcloudConfig(key string) (string, error) {
	out, err := exec.Command("gcloud", "config", "get", key).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
