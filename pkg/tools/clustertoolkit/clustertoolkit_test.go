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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestResolveDownloadDir(t *testing.T) {
	baseDir := t.TempDir()

	tests := []struct {
		name    string
		rawDir  string
		want    string
		wantErr string
	}{
		{
			name:   "current directory dot",
			rawDir: ".",
			want:   "cluster-toolkit",
		},
		{
			name:   "current directory dot slash",
			rawDir: "./",
			want:   "cluster-toolkit",
		},
		{
			name:   "relative subdirectory",
			rawDir: "my-toolkit",
			want:   filepath.Join("my-toolkit", "cluster-toolkit"),
		},
		{
			name:   "relative subdirectory with trailing slash",
			rawDir: "my-toolkit/",
			want:   filepath.Join("my-toolkit", "cluster-toolkit"),
		},
		{
			name:   "relative path already ending with cluster-toolkit",
			rawDir: filepath.Join("my-toolkit", "cluster-toolkit"),
			want:   filepath.Join("my-toolkit", "cluster-toolkit"),
		},
		{
			name:   "relative path cluster-toolkit with trailing slash",
			rawDir: "cluster-toolkit/",
			want:   "cluster-toolkit",
		},
		{
			name:   "directory name ending in cluster-toolkit suffix is not treated as cluster-toolkit",
			rawDir: "my-cluster-toolkit",
			want:   filepath.Join("my-cluster-toolkit", "cluster-toolkit"),
		},
		{
			name:   "nested relative path",
			rawDir: filepath.Join("a", "b", "c"),
			want:   filepath.Join("a", "b", "c", "cluster-toolkit"),
		},
		{
			name:   "internal traversal staying inside working directory",
			rawDir: filepath.Join("a", "..", "b"),
			want:   filepath.Join("b", "cluster-toolkit"),
		},
		{
			name:    "empty string",
			rawDir:  "",
			wantErr: "cannot be empty",
		},
		{
			name:    "whitespace only",
			rawDir:  "   ",
			wantErr: "cannot be empty",
		},
		{
			name:    "nul byte",
			rawDir:  "foo\x00bar",
			wantErr: "invalid NUL byte",
		},
		{
			name:    "leading dash option",
			rawDir:  "-foo",
			wantErr: "must be a relative path under the current directory",
		},
		{
			name:    "git template option injection with cluster-toolkit suffix",
			rawDir:  "--template=/tmp/evil/cluster-toolkit",
			wantErr: "must be a relative path under the current directory",
		},
		{
			name:    "git upload-pack option injection",
			rawDir:  "--upload-pack=evil",
			wantErr: "must be a relative path under the current directory",
		},
		{
			name:    "double dash",
			rawDir:  "--",
			wantErr: "must be a relative path under the current directory",
		},
		{
			name:    "leading dash hidden behind dot slash",
			rawDir:  "./-foo",
			wantErr: "must be a relative path under the current directory",
		},
		{
			name:    "leading dash hidden behind traversal",
			rawDir:  filepath.Join("a", "..", "-foo"),
			wantErr: "must be a relative path under the current directory",
		},
		{
			name:    "absolute path",
			rawDir:  "/tmp/cluster-toolkit",
			wantErr: "must be a relative path under the current directory",
		},
		{
			name:    "root path",
			rawDir:  "/",
			wantErr: "must be a relative path under the current directory",
		},
		{
			name:    "parent directory",
			rawDir:  "..",
			wantErr: "must be a relative path under the current directory",
		},
		{
			name:    "parent directory escape",
			rawDir:  "../outside/cluster-toolkit",
			wantErr: "must be a relative path under the current directory",
		},
		{
			name:    "nested parent directory escape",
			rawDir:  filepath.Join("a", "..", "..", "outside"),
			wantErr: "must be a relative path under the current directory",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := resolveDownloadDir(baseDir, tc.rawDir)
			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("resolveDownloadDir(%q) error = nil, want substring %q", tc.rawDir, tc.wantErr)
				}
				if !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("resolveDownloadDir(%q) error = %q, want substring %q", tc.rawDir, err.Error(), tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("resolveDownloadDir(%q) unexpected error: %v", tc.rawDir, err)
			}
			if got != tc.want {
				t.Errorf("resolveDownloadDir(%q) = %q, want %q", tc.rawDir, got, tc.want)
			}
		})
	}
}

func TestResolveDownloadDir_SymlinksAndFiles(t *testing.T) {
	baseDir := t.TempDir()
	outsideDir := t.TempDir()

	if err := os.Mkdir(filepath.Join(outsideDir, "existing"), 0o750); err != nil {
		t.Fatalf("mkdir outside existing: %v", err)
	}
	if err := os.Mkdir(filepath.Join(baseDir, "real-sub"), 0o750); err != nil {
		t.Fatalf("mkdir real-sub: %v", err)
	}
	if err := os.Symlink(outsideDir, filepath.Join(baseDir, "escape-link")); err != nil {
		t.Fatalf("symlink escape-link: %v", err)
	}
	if err := os.Symlink(outsideDir, filepath.Join(baseDir, "cluster-toolkit")); err != nil {
		t.Fatalf("symlink cluster-toolkit: %v", err)
	}
	if err := os.Symlink(filepath.Join(outsideDir, "does-not-exist"), filepath.Join(baseDir, "dangling-link")); err != nil {
		t.Fatalf("symlink dangling-link: %v", err)
	}
	if err := os.Symlink(filepath.Join(baseDir, "real-sub"), filepath.Join(baseDir, "internal-link")); err != nil {
		t.Fatalf("symlink internal-link: %v", err)
	}
	if err := os.WriteFile(filepath.Join(baseDir, "regular-file"), []byte("x"), 0o600); err != nil {
		t.Fatalf("write regular-file: %v", err)
	}

	t.Run("symlink escaping baseDir with non-existent child", func(t *testing.T) {
		_, err := resolveDownloadDir(baseDir, "escape-link")
		if err == nil || !strings.Contains(err.Error(), "must resolve within the current directory") {
			t.Fatalf("expected symlink escape error, got: %v", err)
		}
	})

	t.Run("symlink escaping baseDir with existing child", func(t *testing.T) {
		_, err := resolveDownloadDir(baseDir, filepath.Join("escape-link", "existing"))
		if err == nil || !strings.Contains(err.Error(), "must resolve within the current directory") {
			t.Fatalf("expected symlink escape error, got: %v", err)
		}
	})

	t.Run("symlink named cluster-toolkit escaping baseDir when rawDir is dot", func(t *testing.T) {
		_, err := resolveDownloadDir(baseDir, ".")
		if err == nil || !strings.Contains(err.Error(), "must resolve within the current directory") {
			t.Fatalf("expected symlink escape error, got: %v", err)
		}
	})

	t.Run("dangling symlink rejected", func(t *testing.T) {
		_, err := resolveDownloadDir(baseDir, "dangling-link")
		if err == nil || !strings.Contains(err.Error(), "unresolvable path component") {
			t.Fatalf("expected unresolvable path error, got: %v", err)
		}
	})

	t.Run("regular file in path rejected", func(t *testing.T) {
		_, err := resolveDownloadDir(baseDir, "regular-file")
		if err == nil || !strings.Contains(err.Error(), "failed to inspect download_directory") {
			t.Fatalf("expected inspect error, got: %v", err)
		}
	})

	t.Run("internal symlink within baseDir allowed", func(t *testing.T) {
		got, err := resolveDownloadDir(baseDir, "internal-link")
		if err != nil {
			t.Fatalf("unexpected error for internal symlink: %v", err)
		}
		want := filepath.Join("internal-link", "cluster-toolkit")
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

func TestHelperProcess(_ *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	if partialDir := os.Getenv("GO_HELPER_CREATE_DIR"); partialDir != "" {
		if !filepath.IsLocal(partialDir) {
			_, _ = fmt.Fprintf(os.Stderr, "invalid partial dir: %q\n", partialDir)
			os.Exit(1)
		}
		// #nosec G703 -- test helper writes only inside the test's temp working directory.
		if err := os.MkdirAll(partialDir, 0o750); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "mkdir failed: %v\n", err)
			os.Exit(1)
		}
		// #nosec G703 -- test helper writes only inside the test's temp working directory.
		if err := os.WriteFile(filepath.Join(partialDir, "partial.tmp"), []byte("partial"), 0o600); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "write failed: %v\n", err)
			os.Exit(1)
		}
	}
	if os.Getenv("GO_HELPER_FAIL") == "1" {
		_, _ = fmt.Fprintln(os.Stderr, "simulated git failure")
		os.Exit(1)
	}
	_, _ = fmt.Fprintln(os.Stdout, "Cloning into 'cluster-toolkit'...")
	os.Exit(0)
}

func stubExecCommand(t *testing.T, fail bool, createPartial bool, gotName *string, gotArgs *[]string) {
	t.Helper()
	orig := execCommand
	t.Cleanup(func() {
		execCommand = orig
	})
	execCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		*gotName = name
		*gotArgs = append([]string(nil), args...)
		cmdArgs := []string{"-test.run=TestHelperProcess", "--"}
		// #nosec G204,G702 -- re-execs the test binary itself with fixed helper arguments.
		cmd := exec.CommandContext(ctx, os.Args[0], cmdArgs...)
		cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
		if fail {
			cmd.Env = append(cmd.Env, "GO_HELPER_FAIL=1")
		}
		if createPartial && len(args) > 0 {
			cmd.Env = append(cmd.Env, "GO_HELPER_CREATE_DIR="+args[len(args)-1])
		}
		return cmd
	}
}

func TestClusterToolkitDownload(t *testing.T) {
	workDir := t.TempDir()
	t.Chdir(workDir)
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() failed: %v", err)
	}

	t.Run("valid relative directory returns resolved path and passes end-of-options separator", func(t *testing.T) {
		var gotName string
		var gotArgs []string
		stubExecCommand(t, false, false, &gotName, &gotArgs)

		res, _, err := clusterToolkitDownload(context.Background(), &mcp.CallToolRequest{}, &clusterToolkitDownloadArgs{
			DownloadDirectory: "my-toolkit",
		})
		if err != nil {
			t.Fatalf("clusterToolkitDownload() unexpected error: %v", err)
		}
		if gotName != "git" {
			t.Errorf("command = %q, want %q", gotName, "git")
		}
		wantRel := filepath.Join("my-toolkit", "cluster-toolkit")
		wantArgs := []string{"clone", clusterToolkitRepoURL, "--", wantRel}
		if !slices.Equal(gotArgs, wantArgs) {
			t.Errorf("args = %v, want %v", gotArgs, wantArgs)
		}
		if len(res.Content) != 1 {
			t.Fatalf("len(Content) = %d, want 1", len(res.Content))
		}
		text, ok := res.Content[0].(*mcp.TextContent)
		if !ok {
			t.Fatalf("unexpected content type: %T", res.Content[0])
		}
		wantDest := filepath.Join(cwd, wantRel)
		if !strings.Contains(text.Text, wantDest) {
			t.Errorf("response %q missing resolved path %q", text.Text, wantDest)
		}
		if !strings.Contains(text.Text, "Cloning into") {
			t.Errorf("response %q missing git output", text.Text)
		}
	})

	t.Run("nil args rejected without invoking git", func(t *testing.T) {
		var gotName string
		var gotArgs []string
		stubExecCommand(t, false, false, &gotName, &gotArgs)

		_, _, err := clusterToolkitDownload(context.Background(), &mcp.CallToolRequest{}, nil)
		if err == nil {
			t.Fatal("expected error for nil args, got nil")
		}
		if gotName != "" {
			t.Errorf("execCommand should not be called, got %q", gotName)
		}
	})

	t.Run("malicious paths rejected without invoking git", func(t *testing.T) {
		for _, badDir := range []string{
			"/tmp/vulnhunt-pwned/cluster-toolkit",
			"--template=/tmp/vh-tpl-cluster-toolkit",
			"../escape/cluster-toolkit",
			"./-c",
		} {
			var gotName string
			var gotArgs []string
			stubExecCommand(t, false, false, &gotName, &gotArgs)

			_, _, err := clusterToolkitDownload(context.Background(), &mcp.CallToolRequest{}, &clusterToolkitDownloadArgs{
				DownloadDirectory: badDir,
			})
			if err == nil {
				t.Errorf("expected error for %q, got nil", badDir)
			}
			if gotName != "" {
				t.Errorf("execCommand should not be called for %q, got %q %v", badDir, gotName, gotArgs)
			}
		}
	})

	t.Run("git clone error includes output and cleans up partial directory", func(t *testing.T) {
		var gotName string
		var gotArgs []string
		stubExecCommand(t, true, true, &gotName, &gotArgs)

		_, _, err := clusterToolkitDownload(context.Background(), &mcp.CallToolRequest{}, &clusterToolkitDownloadArgs{
			DownloadDirectory: "partial-parent",
		})
		if err == nil {
			t.Fatal("expected error when git fails, got nil")
		}
		if !strings.Contains(err.Error(), "failed to download Cluster Toolkit") {
			t.Errorf("unexpected error prefix: %v", err)
		}
		if !strings.Contains(err.Error(), "simulated git failure") {
			t.Errorf("error %q did not include git output", err.Error())
		}
		partialDest := filepath.Join(cwd, "partial-parent", "cluster-toolkit")
		if _, statErr := os.Lstat(partialDest); !errors.Is(statErr, os.ErrNotExist) {
			t.Errorf("expected partial directory %q to be cleaned up, statErr = %v", partialDest, statErr)
		}
	})

	t.Run("existing destination is preserved when git clone fails", func(t *testing.T) {
		existingDir := filepath.Join(cwd, "existing-parent", "cluster-toolkit")
		if err := os.MkdirAll(existingDir, 0o750); err != nil {
			t.Fatalf("MkdirAll(%q) failed: %v", existingDir, err)
		}
		sentinel := filepath.Join(existingDir, "keep.txt")
		if err := os.WriteFile(sentinel, []byte("keep"), 0o600); err != nil {
			t.Fatalf("WriteFile(%q) failed: %v", sentinel, err)
		}

		var gotName string
		var gotArgs []string
		stubExecCommand(t, true, false, &gotName, &gotArgs)

		_, _, err := clusterToolkitDownload(context.Background(), &mcp.CallToolRequest{}, &clusterToolkitDownloadArgs{
			DownloadDirectory: "existing-parent",
		})
		if err == nil {
			t.Fatal("expected error when git fails, got nil")
		}
		if _, statErr := os.Stat(sentinel); statErr != nil {
			t.Errorf("expected existing file %q to be preserved, got statErr = %v", sentinel, statErr)
		}
	})
}

func TestClusterToolkitDownload_MCPDispatch(t *testing.T) {
	workDir := t.TempDir()
	t.Chdir(workDir)
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() failed: %v", err)
	}

	server := mcp.NewServer(&mcp.Implementation{Name: "test-server", Version: "1.0.0"}, nil)
	if err := Install(context.Background(), server, nil); err != nil {
		t.Fatalf("Install() failed: %v", err)
	}

	clientTransport, serverTransport := mcp.NewInMemoryTransports()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = server.Run(ctx, serverTransport)
	}()

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0.0"}, nil)
	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client.Connect() failed: %v", err)
	}
	defer func() { _ = session.Close() }()

	toolsRes, err := session.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		t.Fatalf("ListTools() failed: %v", err)
	}
	if len(toolsRes.Tools) != 1 {
		t.Fatalf("len(Tools) = %d, want 1", len(toolsRes.Tools))
	}
	schemaBytes, err := json.Marshal(toolsRes.Tools[0].InputSchema)
	if err != nil {
		t.Fatalf("json.Marshal(InputSchema) failed: %v", err)
	}
	if !strings.Contains(string(schemaBytes), "MCP server's current working directory") {
		t.Errorf("InputSchema %s does not clarify MCP server's working directory", schemaBytes)
	}

	for _, badDir := range []string{
		"/tmp/vulnhunt-pwned/cluster-toolkit",
		"--template=/tmp/vh-tpl-cluster-toolkit",
	} {
		var gotName string
		var gotArgs []string
		stubExecCommand(t, false, false, &gotName, &gotArgs)

		res, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name: "cluster_toolkit_download",
			Arguments: map[string]any{
				"download_directory": badDir,
			},
		})
		if err != nil {
			t.Fatalf("CallTool(%q) protocol error: %v", badDir, err)
		}
		if !res.IsError {
			t.Errorf("CallTool(%q) IsError = false, want true", badDir)
		}
		if gotName != "" {
			t.Errorf("execCommand was invoked for %q: %q %v", badDir, gotName, gotArgs)
		}
	}

	var gotName string
	var gotArgs []string
	stubExecCommand(t, false, false, &gotName, &gotArgs)

	res, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "cluster_toolkit_download",
		Arguments: map[string]any{
			"download_directory": "my-toolkit",
		},
	})
	if err != nil {
		t.Fatalf("CallTool(my-toolkit) failed: %v", err)
	}
	if res.IsError {
		t.Fatalf("CallTool(my-toolkit) returned error result: %#v", res.Content)
	}
	wantRel := filepath.Join("my-toolkit", "cluster-toolkit")
	wantArgs := []string{"clone", clusterToolkitRepoURL, "--", wantRel}
	if gotName != "git" || !slices.Equal(gotArgs, wantArgs) {
		t.Errorf("got exec %q %v, want %q %v", gotName, gotArgs, "git", wantArgs)
	}
	if len(res.Content) != 1 {
		t.Fatalf("len(Content) = %d, want 1", len(res.Content))
	}
	text, ok := res.Content[0].(*mcp.TextContent)
	if !ok || !strings.Contains(text.Text, filepath.Join(cwd, wantRel)) {
		t.Errorf("CallTool response %#v missing resolved destination path", res.Content[0])
	}
}
