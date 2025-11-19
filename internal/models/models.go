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

import "github.com/google/go-github/v32/github"

// Platform represents an OS and architecture combination
type Platform struct {
	Os   string `json:"os"`
	Arch string `json:"arch"`
}

// VersionResponse represents the response structure for version listings
type VersionResponse struct {
	ID       string      `json:"id"`
	Versions []Version   `json:"versions"`
	Warnings interface{} `json:"warnings"`
}

// GPGPublicKey represents a GPG public key for signature verification
type GPGPublicKey struct {
	KeyID          string      `json:"key_id"`
	ASCIIArmor     string      `json:"ascii_armor"`
	TrustSignature string      `json:"trust_signature"`
	Source         string      `json:"source"`
	SourceURL      interface{} `json:"source_url"`
}

// SigningKeys represents the signing keys collection
type SigningKeys struct {
	GpgPublicKeys []GPGPublicKey `json:"gpg_public_keys,omitempty"`
}

// DownloadResponse represents the response structure for download requests
type DownloadResponse struct {
	Protocols           []string    `json:"protocols,omitempty"`
	Os                  string      `json:"os"`
	Arch                string      `json:"arch"`
	Filename            string      `json:"filename"`
	DownloadURL         string      `json:"download_url"`
	ShasumsURL          string      `json:"shasums_url"`
	ShasumsSignatureURL string      `json:"shasums_signature_url"`
	Shasum              string      `json:"shasum"`
	SigningKeys         SigningKeys `json:"signing_keys"`
}

// ErrorResponse represents an error response structure
type ErrorResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

// Version represents a provider version with its metadata
type Version struct {
	Version      string               `json:"version"`
	Protocols    []string             `json:"protocols,omitempty"`
	Platforms    []Platform           `json:"platforms"`
	ReleaseAsset *github.ReleaseAsset `json:"-"`
}

// ServiceDiscoveryResponse represents the service discovery response
type ServiceDiscoveryResponse struct {
	Providers string `json:"providers.v1"`
}
