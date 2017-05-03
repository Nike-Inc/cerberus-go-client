// Package auth contains various implementations for authenticating with Cerberus.
// These implementations can be used standalone from the main Cerberus client
// to get a login token or manage authentication without having to set up a full client
package auth

import (
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/Nike-Inc/cerberus-go-client/api"
	"github.com/Nike-Inc/cerberus-go-client/utils"
)

// The Auth interface describes the methods that all authentication providers must satisfy
type Auth interface {
	// GetToken should either return an existing token or perform all authentication steps
	// necessary to get a new token. It takes a file object as an argument as a place to
	// read an OTP for MFA flow
	GetToken(*os.File) (string, error)
	// IsAuthenticated should return whether or not there is a valid token. A valid token
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
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, fmt.Errorf("Problem while performing request to Cerberus: %v", err)
	}
	r, checkErr := utils.CheckAndParse(resp)
	if checkErr != nil {
		return nil, checkErr
	}
	return r, nil
}

// Logout takes a set of headers containing a vault token and a URL and logs out of Cerberus.
func Logout(builtURL url.URL, headers http.Header) error {
	builtURL.Path = "/v1/auth"
	req, err := http.NewRequest("DELETE", builtURL.String(), nil)
	if err != nil {
		return err
	}
	req.Header = headers
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return fmt.Errorf("Problem while performing request to Cerberus: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("Unable to log out. Got HTTP response code %d", resp.StatusCode)
	}
	return nil
}
