// Package utils contains common functionality needed across the Cerberus Go client
package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.nike.com/ngp/cerberus-client-go/api"
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

// CheckAndParse is a helper function to check for user auth and token refresh errors and parse a response. It will return a user friendly error
func CheckAndParse(resp *http.Response) (*api.UserAuthResponse, error) {
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, api.ErrorUnauthorized
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Error while trying to authenticate. Got HTTP response code %d", resp.StatusCode)
	}
	decoder := json.NewDecoder(resp.Body)
	u := &api.UserAuthResponse{}
	err := decoder.Decode(u)
	if err != nil {
		return nil, fmt.Errorf("Error while trying to parse response from Cerberus: %v", err)
	}
	return u, nil
}
