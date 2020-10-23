/*
Copyright 2017 Nike Inc.

Licensed under the Apache License, Version 2.0 (the License);
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an AS IS BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package utils contains common functionality needed across the Cerberus Go client
package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/Nike-Inc/cerberus-go-client/v3/api"
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

var ErrorBodyNotReturned = fmt.Errorf("No error body returned from server")

// utils.ParseAPIError is a helper for parsing an error response body from the API.
// If the body doesn't have an error, it will return ErrorBodyNotReturned to indicate that there was no error body sent (probably means there was a server error)
func ParseAPIError(r io.Reader) error {
	var apiErr = api.ErrorResponse{}
	if err := json.NewDecoder(r).Decode(&apiErr); err != nil {
		// If the body is empty or a string, it will hit this error
		if err == io.EOF {
			return ErrorBodyNotReturned
		}
		return fmt.Errorf("Error while parsing API error response: %v", err)
	}
	// Check to see if there is an error ID set and return a different error if not
	// This is here because if there is a json body, it will parse it as valid and won't error
	if apiErr.ErrorID == "" {
		return ErrorBodyNotReturned
	}
	return apiErr
}


