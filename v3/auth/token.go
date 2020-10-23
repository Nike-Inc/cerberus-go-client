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

package auth

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/Nike-Inc/cerberus-go-client/v3/api"
	"github.com/Nike-Inc/cerberus-go-client/v3/utils"
)

// TokenAuth uses a preexisting token to authenticate to Cerberus
type TokenAuth struct {
	token   string
	headers http.Header
	baseURL *url.URL
}

// NewTokenAuth takes a Cerberus URL and valid token and returns a new TokenAuth.
// There is no checking done on whether or not the token is valid, so the function
// expects the a valid token.
func NewTokenAuth(cerberusURL, token string) (*TokenAuth, error) {
	// Make sure that the passed variables are not empty
	if len(cerberusURL) == 0 {
		return nil, fmt.Errorf("Cerberus URL cannot be empty")
	}
	if len(token) == 0 {
		return nil, fmt.Errorf("Token cannot be empty")
	}

	// Parse the URL
	parsedURL, err := utils.ValidateURL(cerberusURL)
	if err != nil {
		return nil, err
	}
	var headers = http.Header{}
	headers.Set("Content-Type", "application/json")
	headers.Set("Accept", "application/json")
	headers.Set("X-Cerberus-Token", token)
	return &TokenAuth{
		baseURL: parsedURL,
		headers: headers,
		token:   token,
	}, nil
}

// GetToken returns the token passed when creating the TokenAuth. Nil should
// be passed as the argument to the function. The argument exists for compatibility
// with the Auth interface
func (t *TokenAuth) GetToken(f *os.File) (string, error) {
	if !t.IsAuthenticated() {
		return "", api.ErrorUnauthenticated
	}
	return t.token, nil
}

// IsAuthenticated always returns true if there is a token. If Logout has been
// called, it will return false
func (t *TokenAuth) IsAuthenticated() bool {
	return t.token != ""
}

// Refresh attempts to refresh the token
func (t *TokenAuth) Refresh() error {
	if !t.IsAuthenticated() {
		return api.ErrorUnauthenticated
	}
	r, err := Refresh(*t.baseURL, t.headers)
	if err != nil {
		return err
	}
	t.token = r.Data.ClientToken.ClientToken
	t.headers.Set("X-Cerberus-Token", r.Data.ClientToken.ClientToken)
	return nil
}

// Logout logs the current token out and removes it from the authentication type
func (t *TokenAuth) Logout() error {
	if !t.IsAuthenticated() {
		return api.ErrorUnauthenticated
	}
	// Use a copy of the base URL
	if err := Logout(*t.baseURL, t.headers); err != nil {
		return err
	}
	// Reset the token and header
	t.token = ""
	t.headers.Del("X-Cerberus-Token")
	return nil
}

// GetHeaders returns HTTP headers used for requests if the method is currently authenticated.
// Returns an error otherwise
func (t *TokenAuth) GetHeaders() (http.Header, error) {
	if !t.IsAuthenticated() {
		return nil, api.ErrorUnauthenticated
	}
	return t.headers, nil
}

// GetURL returns the URL for cerberus
func (t *TokenAuth) GetURL() *url.URL {
	return t.baseURL
}

// Always return zero-valued time.Time struct and a non-nil error
func (t *TokenAuth) GetExpiry() (time.Time, error) {
	return time.Time{}, fmt.Errorf("Expiry time not set")
}
