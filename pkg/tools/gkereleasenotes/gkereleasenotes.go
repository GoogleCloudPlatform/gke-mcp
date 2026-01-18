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

package gkereleasenotes

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
	"github.com/PuerkitoBio/goquery"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var (
	gkeVersionRegexp         = regexp.MustCompile(`\d+\.\d+\.\d+-gke\.\d+`)
	releaseDateHeadingRegexp = regexp.MustCompile(`(^|\n)\s*[A-Za-z]+\s+\d+,\s+\d+\s*(\n|$)`)
)

type getGkeReleaseNotesArgs struct {
	SourceVersion string `json:"SourceVersion" jsonschema:"A source GKE version an upgrade happens from. For example, '1.33.5-gke.120000'."`
	TargetVersion string `json:"TargetVersion" jsonschema:"A target GKE version an upgrade happens from. For example, '1.34.3-gke.240500'."`
}

func Install(_ context.Context, s *mcp.Server, _ *config.Config) error {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_gke_release_notes",
		Description: "Get GKE release notes. Prefer to use this tool if GKE release notes are needed.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:   true,
			IdempotentHint: true,
		},
	}, getGkeReleaseNotes)

	return nil
}

func getGkeReleaseNotes(ctx context.Context, req *mcp.CallToolRequest, args *getGkeReleaseNotesArgs) (*mcp.CallToolResult, any, error) {
	releaseNotesFilePath := fmt.Sprintf("release-notes-%s.html", time.Now().Format("2006-01-02"))

	var out []byte
	var err error

	if _, err = os.Stat(releaseNotesFilePath); err == nil {
		log.Printf("Reading release notes from cached file: %s", releaseNotesFilePath)
		out, err = os.ReadFile(releaseNotesFilePath)
		if err != nil {
			log.Printf("Failed to read cached release notes file: %v", err)
			return nil, nil, err
		}
	} else {
		log.Printf("Fetching release notes from web")
		const releaseNotesPageUrl = "https://cloud.google.com/kubernetes-engine/docs/release-notes"
		resp, err := http.Get(releaseNotesPageUrl)
		if err != nil {
			log.Printf("Failed to get release notes: %v", err)
			return nil, nil, err
		}
		defer func() { _ = resp.Body.Close() }()
		out, err = io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Failed to read release notes response body: %v", err)
			return nil, nil, err
		}
		if err = os.WriteFile(releaseNotesFilePath, out, 0644); err != nil {
			log.Printf("Failed to write release notes to file: %v", err)
		}
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(out))
	if err != nil {
		log.Printf("Failed to parse release notes html content: %v", err)

		return nil, nil, err
	}

	var fullReleaseNotesContent strings.Builder
	doc.Find("[data-text$=\"Version updates\"]").Parent().Parent().Remove()
	doc.Find("[data-text$=\"Security updates\"]").Parent().Parent().Remove()
	doc.Find(".releases").Each(func(i int, s *goquery.Selection) {
		fullReleaseNotesContent.WriteString(s.Text())
	})
	fullReleaseNotesContentText := fullReleaseNotesContent.String()

	reducedReleaseNotes, err := extractReleaseNotesRelevantForUpgrade(fullReleaseNotesContentText, args.SourceVersion, args.TargetVersion)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: reducedReleaseNotes},
		},
	}, nil, nil
}

func extractReleaseNotesRelevantForUpgrade(fullReleaseNotes string, sourceVersion string, targetVersion string) (string, error) {
	versionLocations := gkeVersionRegexp.FindAllStringIndex(fullReleaseNotes, -1)

	var leftBorderVersionLocation []int
	var rightBorderVersionLocation []int
	if versionLocations != nil {
		// The release notes are ordered from newest to oldest.
		// Find the first version that is <= targetVersion. One version to the left (if not first) is our left border.
		for locIndex, loc := range versionLocations {
			version := fullReleaseNotes[loc[0]:loc[1]]
			cmp, err := compareVersions(version, targetVersion)
			if err != nil {
				continue // Skip invalid versions
			}
			fmt.Printf("cmp%d + %v vs %v\n", cmp, version, targetVersion)
			// cmp >= 0 means targetVersion >= version
			if cmp == 0 {
				leftBorderVersionLocation = loc
				break
			} else if cmp > 0 {
				if locIndex == 0 {
					leftBorderVersionLocation = loc
				} else {
					leftBorderVersionLocation = versionLocations[locIndex-1]
				}
				break
			}
		}

		// Find the first version that is >= sourceVersion searching from the end. One version to the right (if not last) is our right border.
		for i := range versionLocations {
			iFromEnd := len(versionLocations) - i - 1
			loc := versionLocations[iFromEnd]
			version := fullReleaseNotes[loc[0]:loc[1]]
			cmp, err := compareVersions(version, sourceVersion)
			if err != nil {
				continue // Skip invalid versions
			}
			if cmp == 0 {
				rightBorderVersionLocation = loc
				break
			} else if cmp < 0 {
				if iFromEnd == len(versionLocations)-1 {
					rightBorderVersionLocation = loc
				} else {
					rightBorderVersionLocation = versionLocations[iFromEnd+1]
				}
				break
			}
		}
	}

	leftBorder := 0
	if leftBorderVersionLocation != nil {
		leftBorder = leftBorderVersionLocation[0]
	}
	rightBorder := len(fullReleaseNotes)
	if rightBorderVersionLocation != nil {
		rightBorder = rightBorderVersionLocation[1]
	}
	reducedReleaseNotes := fullReleaseNotes[leftBorder:rightBorder]

	leftAppend := ""
	leftCut := fullReleaseNotes[:leftBorder]
	if len(leftCut) > 0 {
		dateReleaseHeadingLocations := releaseDateHeadingRegexp.FindAllStringIndex(leftCut, -1)
		if dateReleaseHeadingLocations == nil {
			leftAppend = leftCut
		} else {
			lastDateReleaseHeadingLocation := dateReleaseHeadingLocations[len(dateReleaseHeadingLocations)-1]
			leftAppend = leftCut[lastDateReleaseHeadingLocation[0]:]
		}
	}

	rightAppend := ""
	rightCut := fullReleaseNotes[rightBorder:]
	if len(rightCut) > 0 {
		dateReleaseHeadingLocations := releaseDateHeadingRegexp.FindAllStringIndex(rightCut, -1)
		if dateReleaseHeadingLocations == nil {
			rightAppend = rightCut
		} else {
			firstDateReleaseHeadingLocation := dateReleaseHeadingLocations[0]
			rightCutAppendEnd := firstDateReleaseHeadingLocation[0] - 1
			if rightCutAppendEnd < 0 {
				rightCutAppendEnd = 0
			}
			rightAppend = rightCut[:rightCutAppendEnd]
		}
	}

	reducedReleaseNotes = leftAppend + reducedReleaseNotes + rightAppend

	return reducedReleaseNotes, nil

}

// compareVersion returns:
// - 1 if b > a
// - 0 if b == a
// - -1 if b < a
func compareVersions(a, b string) (int, error) {
	a_major, a_minor, a_patch, a_gke, err := parseGkeVersion(a)
	if err != nil {
		log.Printf("Failed to parse version A '%s': %v", a, err)
		return 0, err
	}
	b_major, b_minor, b_patch, b_gke, err := parseGkeVersion(b)
	if err != nil {
		log.Printf("Failed to parse version B '%s': %v", b, err)
		return 0, err
	}

	if b_major > a_major {
		return 1, nil
	} else if b_major < a_major {
		return -1, nil
	}

	if b_minor > a_minor {
		return 1, nil
	} else if b_minor < a_minor {
		return -1, nil
	}

	if b_patch > a_patch {
		return 1, nil
	} else if b_patch < a_patch {
		return -1, nil
	}

	if b_gke > a_gke {
		return 1, nil
	} else if b_gke < a_gke {
		return -1, nil
	}

	return 0, nil
}

// parseGkeVersion returns 4 ints: major, minor, patch and GKE patch versions
func parseGkeVersion(version string) (int, int, int, int, error) {
	parts := strings.Split(version, "-gke.")
	if len(parts) != 2 {
		return 0, 0, 0, 0, fmt.Errorf("invalid GKE version format: %s", version)
	}

	k8sVersionPart := parts[0]
	gkeVersionPart := parts[1]

	k8sParts := strings.Split(k8sVersionPart, ".")
	if len(k8sParts) != 3 {
		return 0, 0, 0, 0, fmt.Errorf("invalid Kubernetes version part in GKE version: %s", k8sVersionPart)
	}

	major, err := strconv.Atoi(k8sParts[0])
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("cannot parse major version: %w", err)
	}
	minor, err := strconv.Atoi(k8sParts[1])
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("cannot parse minor version: %w", err)
	}
	patch, err := strconv.Atoi(k8sParts[2])
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("cannot parse patch version: %w", err)
	}
	gkePatch, err := strconv.Atoi(gkeVersionPart)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("cannot parse GKE patch version: %w", err)
	}
	return major, minor, patch, gkePatch, nil
}
