package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/google/go-github/v32/github"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
	"os"
	"regexp"
	"strings"
)

var (
	shasumRegexp = regexp.MustCompile(`^(?P<provider>[^_]+)_(?P<version>[^_]+)_SHA256SUMS`)
	binaryRegexp = regexp.MustCompile(`^(?P<provider>[^_]+)_(?P<version>[^_]+)_(?P<os>\w+)_(?P<arch>\w+)`)
	actionRegexp = regexp.MustCompile(`^(?P<version>[^/]+)/(?P<action>[^/]+)/(?P<os>[^/]+)/(?P<arch>\w+)`)
	)

func main() {
	registryHost := os.Getenv("REGISTRY_HOST")

	e := echo.New()
	e.Use(middleware.Logger())

	e.GET("/.well-known/terraform.json", serviceDiscoveryHandler(registryHost))
	e.GET("/v1/providers/:namespace/:type/*", providerHandler(registryHost))

	_ = e.Start(":8080")
}

func serviceDiscoveryHandler(registryHost string) echo.HandlerFunc {
	return func(c echo.Context) error {
		response := struct {
			Providers string `json:"providers.v1"`
		}{
			Providers: registryHost+"/v1/providers/",
		}
		return c.JSON(http.StatusOK, response)
	}
}

type Platform struct {
		Os   string `json:"os"`
		Arch string `json:"arch"`
}

type VersionResponse struct {
	ID       string `json:"id"`
	Versions []Version `json:"versions"`
	Warnings interface{} `json:"warnings"`
}

type GPGPublicKey struct {
KeyID          string      `json:"key_id"`
ASCIIArmor     string      `json:"ascii_armor"`
TrustSignature string      `json:"trust_signature"`
Source         string      `json:"source"`
SourceURL      interface{} `json:"source_url"`
}

type DownloadResponse struct {
	Protocols           []string `json:"protocols,omitempty"`
	Os                  string   `json:"os"`
	Arch                string   `json:"arch"`
	Filename            string   `json:"filename"`
	DownloadURL         string   `json:"download_url"`
	ShasumsURL          string   `json:"shasums_url"`
	ShasumsSignatureURL string   `json:"shasums_signature_url"`
	Shasum              string   `json:"shasum"`
	SigningKeys         struct {
		GpgPublicKeys []GPGPublicKey `json:"gpg_public_keys,omitempty"`
	} `json:"signing_keys"`
}

type ErrorResponse struct {
	Status int `json:"status"`
	Message string `json:"message"`
}

type Version struct {
	Version   string   `json:"version"`
	Protocols []string `json:"protocols,omitempty"`
	Platforms []Platform `json:"platforms"`
	ReleaseAsset *github.ReleaseAsset `json:"-"`
}

func getShasum(asset string, shasumURL string) (string, error) {
	resp, err := http.Get(shasumURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
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

func providerHandler(registryHost string) echo.HandlerFunc {
	client := github.NewClient(nil)

	return func(c echo.Context) error {
		namespace := c.Param("namespace")
		typeParam := c.Param("type")
		param := c.Param("*")
		provider := "terraform-provider-"+typeParam

		repos, _, err := client.Repositories.ListReleases(context.Background(),
			namespace, provider, nil)
		if err != nil {
			return c.JSON(http.StatusBadRequest, &ErrorResponse{
				Status: http.StatusBadRequest,
				Message: err.Error(),
			})
		}
		versions, err := parseVersions(repos)
		if err != nil {
			return c.JSON(http.StatusBadRequest, &ErrorResponse{
				Status: http.StatusBadRequest,
				Message: err.Error(),
			})
		}
		switch param {
		case "versions":
			response := &VersionResponse{
				ID: namespace+"/"+typeParam,
				Versions: versions,
			}
			return c.JSON(http.StatusOK, response)
		default:
			return performAction(c, param, provider, repos)
		}
	}
}

func performAction(c echo.Context, param, provider string, repos []*github.RepositoryRelease) error {
	match := actionRegexp.FindStringSubmatch(param)
	fmt.Printf("matching: %s %d\n", param, len(match))
	if len(match) < 2 {
		fmt.Printf("repos: %v\n", repos)
		return c.JSON(http.StatusBadRequest, &ErrorResponse{
			Status: http.StatusBadRequest,
			Message: "invalid request",
		})
	}
	result := make(map[string]string)
	for i, name := range actionRegexp.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}
	version := result["version"]
	os := result["os"]
	arch := result["arch"]
	filename := fmt.Sprintf("%s_%s_%s_%s.zip", provider, version, os, arch)
	shasumFilename := fmt.Sprintf("%s_%s_SHA256SUMS", provider, version)
	shasumSigFilename := fmt.Sprintf("%s_%s_SHA256SUMS.sig", provider, version)
	downloadURL := ""
	shasumURL := ""
	shasumSigURL := ""

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
		c.JSON(http.StatusBadRequest, &ErrorResponse{
			Status: http.StatusBadRequest,
			Message: fmt.Sprintf("cannot find version: %s", version),
		})
	}
	for _, a := range repo.Assets {
		if *a.Name == filename {
			downloadURL = *a.BrowserDownloadURL
			continue
		}
		if *a.Name == shasumFilename {
			shasumURL = *a.BrowserDownloadURL
			continue
		}
		if *a.Name == shasumSigFilename {
			shasumSigURL = *a.BrowserDownloadURL
			continue
		}
	}
	shasum, _ := getShasum(filename, shasumURL)


	switch result["action"] {
	case "download":
		return c.JSON(http.StatusOK, &DownloadResponse{
			Os: result["os"],
			Arch: result["arch"],
			Filename: filename,
			DownloadURL: downloadURL,
			ShasumsSignatureURL: shasumSigURL,
			ShasumsURL: shasumURL,
			Shasum: shasum,
		})
	default:
		return c.JSON(http.StatusBadRequest, &ErrorResponse{
			Status: http.StatusBadRequest,
			Message: fmt.Sprintf("unsupported action %s", result["action"]),
		})
	}
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
		fmt.Printf("result: %v, match: %v\n", result, match)
		platforms = append(platforms, Platform{
			Os: result["os"],
			Arch: result["arch"],
		})
	}
	return platforms
}