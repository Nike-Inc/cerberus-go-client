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

// Package auth contains various implementations for authenticating with Cerberus.
// These implementations can be used standalone from the main Cerberus client
// to get a login token or manage authentication without having to set up a full client
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

// expiryDelta is the amount of time to subtract from the expiry time to compensate for
// network request time and clock skew
const expiryDelta time.Duration = 60 * time.Second

// The Auth interface describes the methods that all authentication providers must satisfy
type Auth interface {
	// GetToken should either return an existing token or perform all authentication steps
	// necessary to get a new token.
	GetToken(*os.File) (string, error)
	//IsAuthenticated should return whether or not there is a valid token. A valid token
	// is one that exists and is not expired
	IsAuthenticated() bool
	// Refresh uses the current valid token to retrieve a new one
	Refresh() error
	// Logout revokes the current token
	Logout() error
	// GetHeaders is a helper for any client using the authentication strategy.
	// It returns a basic set of headers asking for a JSON response and has
	// the authorization header set with the proper token
	GetHeaders() (http.Header, error)
	GetURL() *url.URL
	// GetExpiry either returns the expiry time of an existing token, or a zero-valued
	// time.Time struct and an error if a token doesn't exist
	GetExpiry() (time.Time, error)
}

// Refresh contains logic for refreshing a token against the API. Because
// all tokens can be refreshed this way, it is better to keep this in one place
func Refresh(builtURL url.URL, headers http.Header) (*api.UserAuthResponse, error) {
	builtURL.Path = "/v2/auth/user/refresh"
	req, err := http.NewRequest("GET", builtURL.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header = headers
	resp, err := (utils.NewHttpClient(headers)).Do(req)
	if err != nil {
		return nil, fmt.Errorf("Problem while performing request to Cerberus: %v", err)
	}
	r, checkErr := utils.CheckAndParse(resp)
	if checkErr != nil {
		return nil, checkErr
	}
	return r, nil
}

// Logout takes a set of headers containing a token and a URL and logs out of Cerberus.
func Logout(builtURL url.URL, headers http.Header) error {
	builtURL.Path = "/v1/auth"
	req, err := http.NewRequest("DELETE", builtURL.String(), nil)
	if err != nil {
		return err
	}
	req.Header = headers
	resp, err := (utils.NewHttpClient(headers)).Do(req)
	if err != nil {
		return fmt.Errorf("Problem while performing request to Cerberus: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("Unable to log out. Got HTTP response code %d", resp.StatusCode)
	}
	return nil
}
