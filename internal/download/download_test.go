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
package download

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetShasum(t *testing.T) {
	tests := []struct {
		name       string
		asset      string
		handler    http.HandlerFunc
		wantShasum string
		wantErr    bool
	}{
		{
			name:  "Valid SHASUM file",
			asset: "terraform-provider-aws_1.0.0_linux_amd64.zip",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`abc123def456  terraform-provider-aws_1.0.0_linux_amd64.zip
789012ghi345  terraform-provider-aws_1.0.0_darwin_amd64.zip
`))
			},
			wantShasum: "abc123def456",
			wantErr:    false,
		},
		{
			name:  "Asset not in SHASUM file",
			asset: "terraform-provider-aws_1.0.0_windows_amd64.zip",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`abc123def456  terraform-provider-aws_1.0.0_linux_amd64.zip
789012ghi345  terraform-provider-aws_1.0.0_darwin_amd64.zip
`))
			},
			wantErr: true,
		},
		{
			name:  "404 Not Found",
			asset: "terraform-provider-aws_1.0.0_linux_amd64.zip",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			},
			wantErr: true,
		},
		{
			name:  "Empty SHASUM file",
			asset: "terraform-provider-aws_1.0.0_linux_amd64.zip",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(``))
			},
			wantErr: true,
		},
		{
			name:  "Multiple spaces in SHASUM line",
			asset: "terraform-provider-aws_1.0.0_linux_amd64.zip",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`abc123def456  terraform-provider-aws_1.0.0_linux_amd64.zip
`))
			},
			wantShasum: "abc123def456",
			wantErr:    false,
		},
		{
			name:  "Single space separator (invalid format)",
			asset: "terraform-provider-aws_1.0.0_linux_amd64.zip",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`abc123def456 terraform-provider-aws_1.0.0_linux_amd64.zip
`))
			},
			wantErr: true,
		},
		{
			name:  "Real-world format",
			asset: "terraform-provider-aws_4.67.0_linux_amd64.zip",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef  terraform-provider-aws_4.67.0_darwin_amd64.zip
fedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321  terraform-provider-aws_4.67.0_darwin_arm64.zip
abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890  terraform-provider-aws_4.67.0_linux_amd64.zip
567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef123456  terraform-provider-aws_4.67.0_windows_amd64.zip
`))
			},
			wantShasum: "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			shasum, err := GetShasum(tt.asset, server.URL)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetShasum() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && shasum != tt.wantShasum {
				t.Errorf("GetShasum() = %v, want %v", shasum, tt.wantShasum)
			}
		})
	}
}

func TestGetShasumInvalidURL(t *testing.T) {
	_, err := GetShasum("test.zip", "http://invalid-url-that-does-not-exist.local:99999")
	if err == nil {
		t.Error("GetShasum() expected error for invalid URL, got nil")
	}
}

func TestGetShasumServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	_, err := GetShasum("test.zip", server.URL)
	if err == nil {
		t.Error("GetShasum() expected error for server error, got nil")
	}
}
