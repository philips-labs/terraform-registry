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
package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"

        "github.com/ProtonMail/go-crypto/openpgp"
        "github.com/ProtonMail/go-crypto/openpgp/packet"
        "github.com/ProtonMail/go-crypto/openpgp/armor"

	"golang.org/x/oauth2"

	"github.com/google/go-github/v32/github"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var (
	shasumRegexp = regexp.MustCompile(`^(?P<provider>[^_]+)_(?P<version>[^_]+)_SHA256SUMS`)
	binaryRegexp = regexp.MustCompile(`^(?P<provider>[^_]+)_(?P<version>[^_]+)_(?P<os>\w+)_(?P<arch>\w+)`)
	actionRegexp = regexp.MustCompile(`^(?P<version>[^/]+)/(?P<action>[^/]+)/(?P<os>[^/]+)/(?P<arch>\w+)`)
)

func main() {
	e := echo.New()
	e.Use(middleware.Logger())

	client := newClient()

	e.GET("/.well-known/terraform.json", serviceDiscoveryHandler())
	e.GET("/v1/providers/:namespace/:type/*", client.providerHandler())

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	_ = e.Start(fmt.Sprintf(":%s", port))
}

func serviceDiscoveryHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		response := struct {
			Providers string `json:"providers.v1"`
		}{
			Providers: "/v1/providers/",
		}
		return c.JSON(http.StatusOK, response)
	}
}

type Client struct {
	github        *github.Client
	authenticated bool
	http          *http.Client
}

type Platform struct {
	Os   string `json:"os"`
	Arch string `json:"arch"`
}

type VersionResponse struct {
	ID       string      `json:"id"`
	Versions []Version   `json:"versions"`
	Warnings interface{} `json:"warnings"`
}

type GPGPublicKey struct {
	KeyID          string      `json:"key_id"`
	ASCIIArmor     string      `json:"ascii_armor"`
	TrustSignature string      `json:"trust_signature"`
	Source         string      `json:"source"`
	SourceURL      interface{} `json:"source_url"`
}

type SigningKeys struct {
	GpgPublicKeys []GPGPublicKey `json:"gpg_public_keys,omitempty"`
}

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

type ErrorResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

type Version struct {
	Version      string               `json:"version"`
	Protocols    []string             `json:"protocols,omitempty"`
	Platforms    []Platform           `json:"platforms"`
	ReleaseAsset *github.ReleaseAsset `json:"-"`
}

func newClient() *Client {
	client := &Client{}

	if token, ok := os.LookupEnv("GITHUB_TOKEN"); ok {
		ctx := context.Background()
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		httpClient := oauth2.NewClient(ctx, ts)

		client.github = github.NewClient(httpClient)
		client.http = httpClient
		client.authenticated = true

	} else {
		client.github = github.NewClient(nil)
	}

	return client
}

func (client *Client) getURL(c echo.Context, asset *github.ReleaseAsset) (string, error) {
	if client.authenticated {
		namespace := c.Get("namespace").(string)
		provider := c.Get("provider").(string)

		_, url, err := client.github.Repositories.DownloadReleaseAsset(context.Background(),
			namespace, provider, *asset.ID, nil)
		if err != nil {
			return "", err
		}

		return url, nil
	}

	return *asset.BrowserDownloadURL, nil
}

func getShasum(asset string, shasumURL string) (string, error) {
	resp, err := http.Get(shasumURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("not found")
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), "  ")
		if len(parts) != 2 {
			continue
		}
		if parts[1] == asset {
			return parts[0], nil
		}
	}
	return "", fmt.Errorf("not found")
}

func (client *Client) providerHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		namespace := c.Param("namespace")
		typeParam := c.Param("type")
		param := c.Param("*")
		provider := "terraform-provider-" + typeParam

		repos, _, err := client.github.Repositories.ListReleases(context.Background(),
			namespace, provider, nil)
		if err != nil {
			return c.JSON(http.StatusBadRequest, &ErrorResponse{
				Status:  http.StatusBadRequest,
				Message: err.Error(),
			})
		}
		versions, err := parseVersions(repos)
		if err != nil {
			return c.JSON(http.StatusBadRequest, &ErrorResponse{
				Status:  http.StatusBadRequest,
				Message: err.Error(),
			})
		}
		switch param {
		case "versions":
			response := &VersionResponse{
				ID:       namespace + "/" + typeParam,
				Versions: versions,
			}
			return c.JSON(http.StatusOK, response)
		default:
			c.Set("namespace", namespace)
			c.Set("provider", provider)
			return client.performAction(c, param, repos)
		}
	}
}

func (client *Client) performAction(c echo.Context, param string, repos []*github.RepositoryRelease) error {
	match := actionRegexp.FindStringSubmatch(param)
	if len(match) < 2 {
		fmt.Printf("repos: %v\n", repos)
		return c.JSON(http.StatusBadRequest, &ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "invalid request",
		})
	}
	result := make(map[string]string)
	for i, name := range actionRegexp.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}
	provider := c.Get("provider").(string)
	version := result["version"]
	os := result["os"]
	arch := result["arch"]
	filename := fmt.Sprintf("%s_%s_%s_%s.zip", provider, version, os, arch)
	shasumFilename := fmt.Sprintf("%s_%s_SHA256SUMS", provider, version)
	shasumSigFilename := fmt.Sprintf("%s_%s_SHA256SUMS.sig", provider, version)
	signKeyFilename := "signkey.asc"

	downloadURL := ""
	shasumURL := ""
	shasumSigURL := ""
	signKeyURL := ""

	var repo *github.RepositoryRelease
	for _, r := range repos {
		for _, a := range r.Assets {
			if v, err := detectSHASUM(*a.Name); err == nil && version == v.Version {
				repo = r
				break
			}
		}
	}
	if repo == nil {
		return c.JSON(http.StatusBadRequest, &ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: fmt.Sprintf("cannot find version: %s", version),
		})
	}
	for _, a := range repo.Assets {
		if *a.Name == filename {
			downloadURL, _ = client.getURL(c, a)
			continue
		}
		if *a.Name == shasumFilename {
			shasumURL, _ = client.getURL(c, a)
			continue
		}
		if *a.Name == shasumSigFilename {
			shasumSigURL, _ = client.getURL(c, a)
			continue
		}
		if *a.Name == signKeyFilename {
			signKeyURL, _ = client.getURL(c, a)
		}
	}

	shasum, err := getShasum(filename, shasumURL)
	if err != nil {
		return c.JSON(http.StatusBadRequest, &ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: fmt.Sprintf("failed getting shasum %v", err),
		})
	}
	pgpPublicKey, pgpPublicKeyID, err := getPublicKey(signKeyURL)
	if err != nil {
		return c.JSON(http.StatusBadRequest, &ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: fmt.Sprintf("failed getting pgp keys %v", err),
		})
	}

	switch result["action"] {
	case "download":
		return c.JSON(http.StatusOK, &DownloadResponse{
			Os:                  result["os"],
			Arch:                result["arch"],
			Filename:            filename,
			DownloadURL:         downloadURL,
			ShasumsSignatureURL: shasumSigURL,
			ShasumsURL:          shasumURL,
			Shasum:              shasum,
			SigningKeys: SigningKeys{
				GpgPublicKeys: []GPGPublicKey{
					{
						KeyID:      pgpPublicKeyID,
						ASCIIArmor: pgpPublicKey,
					},
				},
			},
		})
	default:
		return c.JSON(http.StatusBadRequest, &ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: fmt.Sprintf("unsupported action %s", result["action"]),
		})
	}
}

func getPublicKey(url string) (string, string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", "", fmt.Errorf("not found")
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	// PGP
	armored := bytes.NewReader(data)
	block, err := armor.Decode(armored)
	if err != nil {
		return "", "", err
	}
	if block == nil || block.Type != openpgp.PublicKeyType {
		return "", "", fmt.Errorf("not a public key")
	}
	reader := packet.NewReader(block.Body)
	pkt, err := reader.Next()
	if err != nil {
		return "", "", err
	}
	key, _ := pkt.(*packet.PublicKey)

	return string(data), key.KeyIdString(), nil
}

func parseVersions(repos []*github.RepositoryRelease) ([]Version, error) {
	details := make([]Version, 0)
	for _, r := range repos {
		for _, a := range r.Assets {
			assetDetails, err := detectSHASUM(*a.Name)
			if err == nil {
				assetDetails.Platforms = collectPlatforms(r.Assets)
				details = append(details, *assetDetails)
				break
			}
		}
	}
	return details, nil
}

func detectSHASUM(name string) (*Version, error) {
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
	return &Version{
		Version: result["version"],
	}, nil
}

func collectPlatforms(assets []*github.ReleaseAsset) []Platform {
	platforms := make([]Platform, 0)
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
		platforms = append(platforms, Platform{
			Os:   result["os"],
			Arch: result["arch"],
		})
	}
	return platforms
}
