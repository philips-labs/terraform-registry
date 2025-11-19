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
package client

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/go-github/v32/github"
	"github.com/labstack/echo/v4"
)

func TestNewClient(t *testing.T) {
	// Save original env vars
	origToken := os.Getenv("GITHUB_TOKEN")
	origServerURL := os.Getenv("GITHUB_ENTERPRISE_URL")
	origUploadURL := os.Getenv("GITHUB_ENTERPRISE_UPLOADS_URL")

	// Restore after test
	defer func() {
		os.Setenv("GITHUB_TOKEN", origToken)
		os.Setenv("GITHUB_ENTERPRISE_URL", origServerURL)
		os.Setenv("GITHUB_ENTERPRISE_UPLOADS_URL", origUploadURL)
	}()

	tests := []struct {
		name              string
		setupEnv          func()
		wantAuthenticated bool
		wantErr           bool
	}{
		{
			name: "Without authentication",
			setupEnv: func() {
				os.Unsetenv("GITHUB_TOKEN")
				os.Unsetenv("GITHUB_ENTERPRISE_URL")
				os.Unsetenv("GITHUB_ENTERPRISE_UPLOADS_URL")
			},
			wantAuthenticated: false,
			wantErr:           false,
		},
		{
			name: "With authentication token",
			setupEnv: func() {
				os.Setenv("GITHUB_TOKEN", "test-token")
				os.Unsetenv("GITHUB_ENTERPRISE_URL")
				os.Unsetenv("GITHUB_ENTERPRISE_UPLOADS_URL")
			},
			wantAuthenticated: true,
			wantErr:           false,
		},
		{
			name: "With enterprise URL",
			setupEnv: func() {
				os.Unsetenv("GITHUB_TOKEN")
				os.Setenv("GITHUB_ENTERPRISE_URL", "https://github.example.com/api/v3/")
				os.Unsetenv("GITHUB_ENTERPRISE_UPLOADS_URL")
			},
			wantAuthenticated: false,
			wantErr:           false,
		},
		{
			name: "With enterprise URL and token",
			setupEnv: func() {
				os.Setenv("GITHUB_TOKEN", "test-token")
				os.Setenv("GITHUB_ENTERPRISE_URL", "https://github.example.com/api/v3/")
				os.Unsetenv("GITHUB_ENTERPRISE_UPLOADS_URL")
			},
			wantAuthenticated: true,
			wantErr:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupEnv()

			client, err := NewClient()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if client == nil {
					t.Fatal("NewClient() returned nil client")
				}
				if client.Github == nil {
					t.Error("NewClient() Github client is nil")
				}
				if client.Authenticated != tt.wantAuthenticated {
					t.Errorf("NewClient() Authenticated = %v, want %v", client.Authenticated, tt.wantAuthenticated)
				}
			}
		})
	}
}

func TestGetURL(t *testing.T) {
	tests := []struct {
		name          string
		authenticated bool
		setupContext  func(c echo.Context)
		wantURL       string
		wantErr       bool
	}{
		{
			name:          "Unauthenticated - use browser download URL",
			authenticated: false,
			setupContext:  func(c echo.Context) {},
			wantURL:       "https://github.com/test/release/download/v1.0.0/asset.zip",
			wantErr:       false,
		},
		// Note: Testing authenticated flow would require mocking GitHub API calls
		// which is complex without dependency injection. This is a basic test.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				Authenticated: tt.authenticated,
				HTTP:          &http.Client{},
				Github:        github.NewClient(nil),
			}

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			tt.setupContext(c)

			browserURL := "https://github.com/test/release/download/v1.0.0/asset.zip"
			asset := &github.ReleaseAsset{
				BrowserDownloadURL: &browserURL,
			}

			url, err := client.GetURL(c, asset)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !tt.authenticated {
				if url != tt.wantURL {
					t.Errorf("GetURL() = %v, want %v", url, tt.wantURL)
				}
			}
		})
	}
}

func TestClientFields(t *testing.T) {
	// Test that Client has the expected fields
	client := &Client{
		Github:        github.NewClient(nil),
		Authenticated: true,
		HTTP:          &http.Client{},
	}

	if client.Github == nil {
		t.Error("Client.Github should not be nil")
	}
	if !client.Authenticated {
		t.Error("Client.Authenticated should be true")
	}
	if client.HTTP == nil {
		t.Error("Client.HTTP should not be nil")
	}
}
