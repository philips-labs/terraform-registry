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
package models

import (
	"encoding/json"
	"testing"

	"github.com/google/go-github/v32/github"
)

func TestPlatformSerialization(t *testing.T) {
	platform := Platform{
		Os:   "linux",
		Arch: "amd64",
	}

	data, err := json.Marshal(platform)
	if err != nil {
		t.Errorf("Failed to marshal platform: %v", err)
	}

	var unmarshaled Platform
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("Failed to unmarshal platform: %v", err)
	}

	if unmarshaled.Os != platform.Os || unmarshaled.Arch != platform.Arch {
		t.Errorf("Platform mismatch: expected %+v, got %+v", platform, unmarshaled)
	}
}

func TestVersionResponseSerialization(t *testing.T) {
	response := VersionResponse{
		ID: "hashicorp/aws",
		Versions: []Version{
			{
				Version: "1.0.0",
				Platforms: []Platform{
					{Os: "linux", Arch: "amd64"},
					{Os: "darwin", Arch: "amd64"},
				},
			},
		},
	}

	data, err := json.Marshal(response)
	if err != nil {
		t.Errorf("Failed to marshal version response: %v", err)
	}

	var unmarshaled VersionResponse
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("Failed to unmarshal version response: %v", err)
	}

	if unmarshaled.ID != response.ID {
		t.Errorf("ID mismatch: expected %s, got %s", response.ID, unmarshaled.ID)
	}

	if len(unmarshaled.Versions) != len(response.Versions) {
		t.Errorf("Versions count mismatch: expected %d, got %d", len(response.Versions), len(unmarshaled.Versions))
	}
}

func TestDownloadResponseSerialization(t *testing.T) {
	response := DownloadResponse{
		Os:                  "linux",
		Arch:                "amd64",
		Filename:            "terraform-provider-aws_1.0.0_linux_amd64.zip",
		DownloadURL:         "https://example.com/download",
		ShasumsURL:          "https://example.com/shasums",
		ShasumsSignatureURL: "https://example.com/shasums.sig",
		Shasum:              "abc123",
		SigningKeys: SigningKeys{
			GpgPublicKeys: []GPGPublicKey{
				{
					KeyID:      "1234567890ABCDEF",
					ASCIIArmor: "-----BEGIN PGP PUBLIC KEY BLOCK-----",
				},
			},
		},
	}

	data, err := json.Marshal(response)
	if err != nil {
		t.Errorf("Failed to marshal download response: %v", err)
	}

	var unmarshaled DownloadResponse
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("Failed to unmarshal download response: %v", err)
	}

	if unmarshaled.Filename != response.Filename {
		t.Errorf("Filename mismatch: expected %s, got %s", response.Filename, unmarshaled.Filename)
	}

	if unmarshaled.Shasum != response.Shasum {
		t.Errorf("Shasum mismatch: expected %s, got %s", response.Shasum, unmarshaled.Shasum)
	}
}

func TestErrorResponseSerialization(t *testing.T) {
	response := ErrorResponse{
		Status:  400,
		Message: "Bad request",
	}

	data, err := json.Marshal(response)
	if err != nil {
		t.Errorf("Failed to marshal error response: %v", err)
	}

	var unmarshaled ErrorResponse
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("Failed to unmarshal error response: %v", err)
	}

	if unmarshaled.Status != response.Status {
		t.Errorf("Status mismatch: expected %d, got %d", response.Status, unmarshaled.Status)
	}

	if unmarshaled.Message != response.Message {
		t.Errorf("Message mismatch: expected %s, got %s", response.Message, unmarshaled.Message)
	}
}

func TestVersionWithReleaseAsset(t *testing.T) {
	assetName := "terraform-provider-aws_1.0.0_SHA256SUMS"
	asset := &github.ReleaseAsset{
		Name: &assetName,
	}

	version := Version{
		Version:      "1.0.0",
		Protocols:    []string{"5.0"},
		Platforms:    []Platform{{Os: "linux", Arch: "amd64"}},
		ReleaseAsset: asset,
	}

	// Test that ReleaseAsset is excluded from JSON serialization
	data, err := json.Marshal(version)
	if err != nil {
		t.Errorf("Failed to marshal version: %v", err)
	}

	var unmarshaled map[string]interface{}
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("Failed to unmarshal version: %v", err)
	}

	// ReleaseAsset should not be present in JSON due to json:"-" tag
	if _, exists := unmarshaled["ReleaseAsset"]; exists {
		t.Error("ReleaseAsset should not be serialized to JSON")
	}
}

func TestServiceDiscoveryResponseSerialization(t *testing.T) {
	response := ServiceDiscoveryResponse{
		Providers: "/v1/providers/",
	}

	data, err := json.Marshal(response)
	if err != nil {
		t.Errorf("Failed to marshal service discovery response: %v", err)
	}

	var unmarshaled ServiceDiscoveryResponse
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("Failed to unmarshal service discovery response: %v", err)
	}

	if unmarshaled.Providers != response.Providers {
		t.Errorf("Providers mismatch: expected %s, got %s", response.Providers, unmarshaled.Providers)
	}
}
