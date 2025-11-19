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
	"fmt"
	"regexp"

	"terraform-registry/internal/models"

	"github.com/google/go-github/v32/github"
)

var (
	shasumRegexp = regexp.MustCompile(`^(?P<provider>[^_]+)_(?P<version>[^_]+)_SHA256SUMS`)
	binaryRegexp = regexp.MustCompile(`^(?P<provider>[^_]+)_(?P<version>[^_]+)_(?P<os>\w+)_(?P<arch>\w+)`)
	ActionRegexp = regexp.MustCompile(`^(?P<version>[^/]+)/(?P<action>[^/]+)/(?P<os>[^/]+)/(?P<arch>\w+)`)
)

// ParseVersions extracts version information from GitHub releases
func ParseVersions(repos []*github.RepositoryRelease) ([]models.Version, error) {
	details := make([]models.Version, 0)
	for _, r := range repos {
		for _, a := range r.Assets {
			assetDetails, err := DetectSHASUM(*a.Name)
			if err == nil {
				assetDetails.Platforms = CollectPlatforms(r.Assets)
				details = append(details, *assetDetails)
				break
			}
		}
	}
	return details, nil
}

// DetectSHASUM detects version information from SHASUM filename
func DetectSHASUM(name string) (*models.Version, error) {
	match := shasumRegexp.FindStringSubmatch(name)
	if len(match) < 2 {
		return nil, fmt.Errorf("nomatch %d", len(match))
	}
	result := make(map[string]string)
	for i, name := range shasumRegexp.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}
	return &models.Version{
		Version: result["version"],
	}, nil
}

// CollectPlatforms extracts platform information from release assets
func CollectPlatforms(assets []*github.ReleaseAsset) []models.Platform {
	platforms := make([]models.Platform, 0)
	for _, a := range assets {
		match := binaryRegexp.FindStringSubmatch(*a.Name)
		if len(match) < 2 {
			continue
		}
		result := make(map[string]string)
		for i, name := range binaryRegexp.SubexpNames() {
			if i != 0 && name != "" {
				result[name] = match[i]
			}
		}
		platforms = append(platforms, models.Platform{
			Os:   result["os"],
			Arch: result["arch"],
		})
	}
	return platforms
}
