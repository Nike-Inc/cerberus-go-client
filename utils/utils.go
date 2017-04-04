// Package utils contains common functionality needed across the Cerberus Go client
package utils

import (
	"fmt"
	"net/url"
)

// ValidateURL takes a cerberus URL and makes sure that it is valid.
// It expects
func ValidateURL(fullURL string) (*url.URL, error) {
	parsed, err := url.Parse(fullURL)
	if err != nil {
		return nil, err
	}
	// Make sure they didn't pass other things
	if parsed.Path != "" {
		return nil, fmt.Errorf("Given URL contained a path: %s. The URL should not have a path", parsed.Path)
	}
	if parsed.RawQuery != "" {
		return nil, fmt.Errorf("Given URL contained a query string: %s. The URL should not have a query string", parsed.RawQuery)
	}
	return parsed, nil
}
