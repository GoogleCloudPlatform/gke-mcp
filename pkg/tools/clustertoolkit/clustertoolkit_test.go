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

package clustertoolkit

import (
	"testing"
)

func TestClusterToolkitDownloadArgs_Fields(t *testing.T) {
	args := clusterToolkitDownloadArgs{
		DownloadDirectory: "/tmp/cluster-toolkit",
	}

	if args.DownloadDirectory != "/tmp/cluster-toolkit" {
		t.Errorf("DownloadDirectory = %s, want /tmp/cluster-toolkit", args.DownloadDirectory)
	}
}

func TestClusterToolkitDownloadArgs_Empty(t *testing.T) {
	args := clusterToolkitDownloadArgs{}
	if args.DownloadDirectory != "" {
		t.Errorf("Expected empty DownloadDirectory, got %s", args.DownloadDirectory)
	}
}

func TestClusterToolkitDownloadArgs_WithTrailingSlash(t *testing.T) {
	args := clusterToolkitDownloadArgs{
		DownloadDirectory: "/home/user/",
	}

	if args.DownloadDirectory != "/home/user/" {
		t.Error("DownloadDirectory should preserve trailing slash")
	}
}

func TestClusterToolkitDownloadArgs_CustomPath(t *testing.T) {
	args := clusterToolkitDownloadArgs{
		DownloadDirectory: "/opt/cluster-toolkit-download",
	}

	if args.DownloadDirectory != "/opt/cluster-toolkit-download" {
		t.Errorf("DownloadDirectory = %s, want /opt/cluster-toolkit-download", args.DownloadDirectory)
	}
}

func TestClusterToolkitDownloadArgs_RelativePath(t *testing.T) {
	args := clusterToolkitDownloadArgs{
		DownloadDirectory: "./downloads",
	}

	if args.DownloadDirectory != "./downloads" {
		t.Errorf("DownloadDirectory = %s, want ./downloads", args.DownloadDirectory)
	}
}

func TestClusterToolkitDownloadArgs_JSONTags(t *testing.T) {
	args := clusterToolkitDownloadArgs{
		DownloadDirectory: "/tmp/test",
	}

	if args.DownloadDirectory != "/tmp/test" {
		t.Error("DownloadDirectory field not working correctly")
	}
}
