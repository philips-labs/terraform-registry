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
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/google/go-github/v32/github"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
)

// Client wraps the GitHub client and provides methods for interacting with GitHub repositories
type Client struct {
	Github        *github.Client
	Authenticated bool
	HTTP          *http.Client
}

// NewClient creates a new Client instance with optional GitHub authentication
func NewClient() (*Client, error) {
	client := &Client{}

	if token, ok := os.LookupEnv("GITHUB_TOKEN"); ok {
		ctx := context.Background()
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)

		client.HTTP = oauth2.NewClient(ctx, ts)
		client.Authenticated = true
	}

	if serverURL := os.Getenv("GITHUB_ENTERPRISE_URL"); serverURL != "" {
		uploadURL := serverURL

		if url := os.Getenv("GITHUB_ENTERPRISE_UPLOADS_URL"); url != "" {
			uploadURL = url
		}

		ghClient, err := github.NewEnterpriseClient(serverURL, uploadURL, client.HTTP)
		if err != nil {
			return nil, fmt.Errorf("could not create enterprise client: %w", err)
		}

		client.Github = ghClient
	} else {
		client.Github = github.NewClient(client.HTTP)
	}

	return client, nil
}

// GetURL retrieves the download URL for a GitHub release asset
func (client *Client) GetURL(c echo.Context, asset *github.ReleaseAsset) (string, error) {
	if client.Authenticated {
		namespace := c.Get("namespace").(string)
		provider := c.Get("provider").(string)

		_, url, err := client.Github.Repositories.DownloadReleaseAsset(context.Background(),
			namespace, provider, *asset.ID, nil)
		if err != nil {
			return "", err
		}

		return url, nil
	}

	return *asset.BrowserDownloadURL, nil
}
