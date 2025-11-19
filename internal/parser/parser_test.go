/*
Copyright © 2020 Andy Lo-A-Foe

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package parser

import (
	"testing"

	"github.com/google/go-github/v32/github"
)

func TestDetectSHASUM(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		wantVersion string
		wantErr     bool
	}{
		{
			name:        "Valid SHASUM filename",
			filename:    "terraform-provider-aws_1.0.0_SHA256SUMS",
			wantVersion: "1.0.0",
			wantErr:     false,
		},
		{
			name:        "Valid SHASUM with complex version",
			filename:    "terraform-provider-kubernetes_2.10.0-beta.1_SHA256SUMS",
			wantVersion: "2.10.0-beta.1",
			wantErr:     false,
		},
		{
			name:     "Invalid filename",
			filename: "terraform-provider-aws.zip",
			wantErr:  true,
		},
		{
			name:     "Empty filename",
			filename: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, err := DetectSHASUM(tt.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("DetectSHASUM() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && version.Version != tt.wantVersion {
				t.Errorf("DetectSHASUM() version = %v, want %v", version.Version, tt.wantVersion)
			}
		})
	}
}

func TestCollectPlatforms(t *testing.T) {
	tests := []struct {
		name          string
		assets        []*github.ReleaseAsset
		wantPlatforms int
	}{
		{
			name: "Multiple platforms",
			assets: []*github.ReleaseAsset{
				{Name: stringPtr("terraform-provider-aws_1.0.0_linux_amd64.zip")},
				{Name: stringPtr("terraform-provider-aws_1.0.0_darwin_amd64.zip")},
				{Name: stringPtr("terraform-provider-aws_1.0.0_windows_amd64.zip")},
				{Name: stringPtr("terraform-provider-aws_1.0.0_SHA256SUMS")},
			},
			wantPlatforms: 3,
		},
		{
			name: "Single platform",
			assets: []*github.ReleaseAsset{
				{Name: stringPtr("terraform-provider-aws_1.0.0_linux_amd64.zip")},
			},
			wantPlatforms: 1,
		},
		{
			name: "No binary assets",
			assets: []*github.ReleaseAsset{
				{Name: stringPtr("terraform-provider-aws_1.0.0_SHA256SUMS")},
				{Name: stringPtr("terraform-provider-aws_1.0.0_SHA256SUMS.sig")},
			},
			wantPlatforms: 0,
		},
		{
			name:          "Empty assets",
			assets:        []*github.ReleaseAsset{},
			wantPlatforms: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			platforms := CollectPlatforms(tt.assets)
			if len(platforms) != tt.wantPlatforms {
				t.Errorf("CollectPlatforms() got %d platforms, want %d", len(platforms), tt.wantPlatforms)
			}
		})
	}
}

func TestCollectPlatformsContent(t *testing.T) {
	assets := []*github.ReleaseAsset{
		{Name: stringPtr("terraform-provider-aws_1.0.0_linux_amd64.zip")},
		{Name: stringPtr("terraform-provider-aws_1.0.0_darwin_arm64.zip")},
	}

	platforms := CollectPlatforms(assets)

	if len(platforms) != 2 {
		t.Fatalf("Expected 2 platforms, got %d", len(platforms))
	}

	// Verify first platform
	if platforms[0].Os != "linux" || platforms[0].Arch != "amd64" {
		t.Errorf("First platform: expected linux/amd64, got %s/%s", platforms[0].Os, platforms[0].Arch)
	}

	// Verify second platform
	if platforms[1].Os != "darwin" || platforms[1].Arch != "arm64" {
		t.Errorf("Second platform: expected darwin/arm64, got %s/%s", platforms[1].Os, platforms[1].Arch)
	}
}

func TestParseVersions(t *testing.T) {
	tests := []struct {
		name         string
		releases     []*github.RepositoryRelease
		wantVersions int
		wantErr      bool
	}{
		{
			name: "Single release with platforms",
			releases: []*github.RepositoryRelease{
				{
					Assets: []*github.ReleaseAsset{
						{Name: stringPtr("terraform-provider-aws_1.0.0_SHA256SUMS")},
						{Name: stringPtr("terraform-provider-aws_1.0.0_linux_amd64.zip")},
						{Name: stringPtr("terraform-provider-aws_1.0.0_darwin_amd64.zip")},
					},
				},
			},
			wantVersions: 1,
			wantErr:      false,
		},
		{
			name: "Multiple releases",
			releases: []*github.RepositoryRelease{
				{
					Assets: []*github.ReleaseAsset{
						{Name: stringPtr("terraform-provider-aws_1.0.0_SHA256SUMS")},
						{Name: stringPtr("terraform-provider-aws_1.0.0_linux_amd64.zip")},
					},
				},
				{
					Assets: []*github.ReleaseAsset{
						{Name: stringPtr("terraform-provider-aws_2.0.0_SHA256SUMS")},
						{Name: stringPtr("terraform-provider-aws_2.0.0_linux_amd64.zip")},
					},
				},
			},
			wantVersions: 2,
			wantErr:      false,
		},
		{
			name: "Release without SHASUM",
			releases: []*github.RepositoryRelease{
				{
					Assets: []*github.ReleaseAsset{
						{Name: stringPtr("terraform-provider-aws_1.0.0_linux_amd64.zip")},
					},
				},
			},
			wantVersions: 0,
			wantErr:      false,
		},
		{
			name:         "Empty releases",
			releases:     []*github.RepositoryRelease{},
			wantVersions: 0,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			versions, err := ParseVersions(tt.releases)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseVersions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(versions) != tt.wantVersions {
				t.Errorf("ParseVersions() got %d versions, want %d", len(versions), tt.wantVersions)
			}
		})
	}
}

func TestActionRegexp(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantMatch   bool
		wantVersion string
		wantAction  string
		wantOs      string
		wantArch    string
	}{
		{
			name:        "Valid download action",
			input:       "1.0.0/download/linux/amd64",
			wantMatch:   true,
			wantVersion: "1.0.0",
			wantAction:  "download",
			wantOs:      "linux",
			wantArch:    "amd64",
		},
		{
			name:        "Valid with complex version",
			input:       "2.10.0-beta.1/download/darwin/arm64",
			wantMatch:   true,
			wantVersion: "2.10.0-beta.1",
			wantAction:  "download",
			wantOs:      "darwin",
			wantArch:    "arm64",
		},
		{
			name:      "Invalid format",
			input:     "versions",
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match := ActionRegexp.FindStringSubmatch(tt.input)
			if (len(match) >= 2) != tt.wantMatch {
				t.Errorf("ActionRegexp match = %v, wantMatch %v", len(match) >= 2, tt.wantMatch)
				return
			}

			if tt.wantMatch {
				result := make(map[string]string)
				for i, name := range ActionRegexp.SubexpNames() {
					if i != 0 && name != "" {
						result[name] = match[i]
					}
				}

				if result["version"] != tt.wantVersion {
					t.Errorf("version = %v, want %v", result["version"], tt.wantVersion)
				}
				if result["action"] != tt.wantAction {
					t.Errorf("action = %v, want %v", result["action"], tt.wantAction)
				}
				if result["os"] != tt.wantOs {
					t.Errorf("os = %v, want %v", result["os"], tt.wantOs)
				}
				if result["arch"] != tt.wantArch {
					t.Errorf("arch = %v, want %v", result["arch"], tt.wantArch)
				}
			}
		})
	}
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
