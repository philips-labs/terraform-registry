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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"terraform-registry/internal/client"
	"terraform-registry/internal/models"

	"github.com/google/go-github/v32/github"
	"github.com/labstack/echo/v4"
)

func TestServiceDiscoveryHandler(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/.well-known/terraform.json", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := ServiceDiscoveryHandler()
	err := handler(c)

	if err != nil {
		t.Errorf("ServiceDiscoveryHandler() error = %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	var response models.ServiceDiscoveryResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	expectedProviders := "/v1/providers/"
	if response.Providers != expectedProviders {
		t.Errorf("Expected Providers %s, got %s", expectedProviders, response.Providers)
	}
}

func TestProviderHandlerStructure(t *testing.T) {
	// Test that the handler is properly structured
	client := &client.Client{
		Github:        github.NewClient(nil),
		Authenticated: false,
	}

	handler := ProviderHandler(client)
	if handler == nil {
		t.Error("ProviderHandler() returned nil")
	}
}

func TestErrorResponseFormat(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Simulate an error response
	err := c.JSON(http.StatusBadRequest, &models.ErrorResponse{
		Status:  http.StatusBadRequest,
		Message: "test error",
	})

	if err != nil {
		t.Errorf("Failed to create error response: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var response models.ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to unmarshal error response: %v", err)
	}

	if response.Status != http.StatusBadRequest {
		t.Errorf("Expected error status %d, got %d", http.StatusBadRequest, response.Status)
	}

	if response.Message != "test error" {
		t.Errorf("Expected error message 'test error', got '%s'", response.Message)
	}
}
