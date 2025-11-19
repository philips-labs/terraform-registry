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
	"bufio"
	"fmt"
	"net/http"
	"strings"
)

// GetShasum retrieves the SHA256 sum for a specific asset from a SHASUM file URL
func GetShasum(asset string, shasumURL string) (string, error) {
	resp, err := http.Get(shasumURL)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

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
