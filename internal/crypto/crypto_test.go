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
package crypto

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Mock PGP public key for testing
const testPublicKey = `-----BEGIN PGP PUBLIC KEY BLOCK-----

mQENBGLsLg4BCADQvqPQUwXz/HT2JMGMTKtdUFwCedrQP3d6ywZHBaGzKGl3xGKk
qTQHHlMv0jKmH5aVaFw9i0pKl9OQGm9pdxL4S+C+zMxmC8QLPMvPjNPLCBsQWMbT
uLqLWXrDmEKJl0MK7f6B3xN2q+LdzGYVgMHxEaRwPK+W8fOHGEZgQvpO6pVe5L5H
sEPx4n5aMjBs5TmzsPT9FrRf7L5fkPPCNJwLEUvGGxOWDdzH8gLqNkVVZxY0HQsd
DpP0Jv3aQpHRGBe0q7xNhqVb6bXLBm3lLUfDxqLkAJqXhVL6xTqWvYQYvJMdYhgL
FLbq9jLdXhTdGvgXKQ4PiEwPJX9nHxLZBAhVABEBAAG0K0hhc2hpQ29ycCBTZWN1
cml0eSA8c2VjdXJpdHlAaGFzaGljb3JwLmNvbT6JAVQEEwEIAD4WIQTgLaKPjc8g
3zQ3gDJvVDcGCLFbOwUCYuwuDgIbAwUJA8JnAAULCQgHAgYVCgkICwIEFgIDAQIe
AQIXgAAKCRBvVDcGCLFbO5R1B/4zlOCckn0sMvL6BL8tLRU8t6w3xD5vqNq7XKQT
lNj8GYHH7BsMEL3VbLmRVb5PjqGbQthXXfOQkVQV7vOaQP8rFp9DJMdC2j8B6Lqx
rF4jQYOPvVvGx5wCJtKyxU8PN8qjPZJzqXQ0jKdLpKTl/yqFBVF0MN8V/6kFAGql
=EXAMPLE
-----END PGP PUBLIC KEY BLOCK-----`

func TestGetPublicKey(t *testing.T) {
	tests := []struct {
		name       string
		handler    http.HandlerFunc
		wantErr    bool
		checkKeyID bool
	}{
		{
			name: "Valid PGP key",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(testPublicKey))
			},
			wantErr:    false,
			checkKeyID: true,
		},
		{
			name: "404 Not Found",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			},
			wantErr: true,
		},
		{
			name: "Invalid PGP data",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("not a valid pgp key"))
			},
			wantErr: true,
		},
		{
			name: "Empty response",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(""))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			armor, keyID, err := GetPublicKey(server.URL)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPublicKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if armor == "" {
					t.Error("GetPublicKey() returned empty armor")
				}
				if tt.checkKeyID && keyID == "" {
					t.Error("GetPublicKey() returned empty keyID")
				}
				if !strings.Contains(armor, "BEGIN PGP PUBLIC KEY BLOCK") {
					t.Error("GetPublicKey() armor doesn't contain PGP header")
				}
			}
		})
	}
}

func TestGetPublicKeyInvalidURL(t *testing.T) {
	_, _, err := GetPublicKey("http://invalid-url-that-does-not-exist.local:99999")
	if err == nil {
		t.Error("GetPublicKey() expected error for invalid URL, got nil")
	}
}

func TestGetPublicKeyNonPGPContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Return valid armored data but not a public key
		w.Write([]byte(`-----BEGIN PGP MESSAGE-----
Some message content here
-----END PGP MESSAGE-----`))
	}))
	defer server.Close()

	_, _, err := GetPublicKey(server.URL)
	if err == nil {
		t.Error("GetPublicKey() expected error for non-public-key content, got nil")
	}
}
