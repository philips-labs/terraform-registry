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
package handler

import (
	"context"
	"fmt"
	"net/http"

	"terraform-registry/internal/client"
	"terraform-registry/internal/crypto"
	"terraform-registry/internal/download"
	"terraform-registry/internal/models"
	"terraform-registry/internal/parser"

	"github.com/google/go-github/v32/github"
	"github.com/labstack/echo/v4"
)

// ServiceDiscoveryHandler returns the service discovery handler
func ServiceDiscoveryHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		response := models.ServiceDiscoveryResponse{
			Providers: "/v1/providers/",
		}
		return c.JSON(http.StatusOK, response)
	}
}

// ProviderHandler returns the provider handler for the given client
func ProviderHandler(client *client.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		namespace := c.Param("namespace")
		typeParam := c.Param("type")
		param := c.Param("*")
		provider := "terraform-provider-" + typeParam

		repos, _, err := client.Github.Repositories.ListReleases(context.Background(),
			namespace, provider, nil)
		if err != nil {
			return c.JSON(http.StatusBadRequest, &models.ErrorResponse{
				Status:  http.StatusBadRequest,
				Message: err.Error(),
			})
		}
		versions, err := parser.ParseVersions(repos)
		if err != nil {
			return c.JSON(http.StatusBadRequest, &models.ErrorResponse{
				Status:  http.StatusBadRequest,
				Message: err.Error(),
			})
		}
		switch param {
		case "versions":
			response := &models.VersionResponse{
				ID:       namespace + "/" + typeParam,
				Versions: versions,
			}
			return c.JSON(http.StatusOK, response)
		default:
			c.Set("namespace", namespace)
			c.Set("provider", provider)
			return performAction(client, c, param, repos)
		}
	}
}

func performAction(client *client.Client, c echo.Context, param string, repos []*github.RepositoryRelease) error {
	match := parser.ActionRegexp.FindStringSubmatch(param)
	if len(match) < 2 {
		fmt.Printf("repos: %v\n", repos)
		return c.JSON(http.StatusBadRequest, &models.ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "invalid request",
		})
	}
	result := make(map[string]string)
	for i, name := range parser.ActionRegexp.SubexpNames() {
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
			if v, err := parser.DetectSHASUM(*a.Name); err == nil && version == v.Version {
				repo = r
				break
			}
		}
	}
	if repo == nil {
		return c.JSON(http.StatusBadRequest, &models.ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: fmt.Sprintf("cannot find version: %s", version),
		})
	}
	for _, a := range repo.Assets {
		if *a.Name == filename {
			downloadURL, _ = client.GetURL(c, a)
			continue
		}
		if *a.Name == shasumFilename {
			shasumURL, _ = client.GetURL(c, a)
			continue
		}
		if *a.Name == shasumSigFilename {
			shasumSigURL, _ = client.GetURL(c, a)
			continue
		}
		if *a.Name == signKeyFilename {
			signKeyURL, _ = client.GetURL(c, a)
		}
	}

	shasum, err := download.GetShasum(filename, shasumURL)
	if err != nil {
		return c.JSON(http.StatusBadRequest, &models.ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: fmt.Sprintf("failed getting shasum %v", err),
		})
	}
	pgpPublicKey, pgpPublicKeyID, err := crypto.GetPublicKey(signKeyURL)
	if err != nil {
		return c.JSON(http.StatusBadRequest, &models.ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: fmt.Sprintf("failed getting pgp keys %v", err),
		})
	}

	switch result["action"] {
	case "download":
		return c.JSON(http.StatusOK, &models.DownloadResponse{
			Os:                  result["os"],
			Arch:                result["arch"],
			Filename:            filename,
			DownloadURL:         downloadURL,
			ShasumsSignatureURL: shasumSigURL,
			ShasumsURL:          shasumURL,
			Shasum:              shasum,
			SigningKeys: models.SigningKeys{
				GpgPublicKeys: []models.GPGPublicKey{
					{
						KeyID:      pgpPublicKeyID,
						ASCIIArmor: pgpPublicKey,
					},
				},
			},
		})
	default:
		return c.JSON(http.StatusBadRequest, &models.ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: fmt.Sprintf("unsupported action %s", result["action"]),
		})
	}
}
