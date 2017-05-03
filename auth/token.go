package auth

import (
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/Nike-Inc/cerberus-go-client/api"
	"github.com/Nike-Inc/cerberus-go-client/utils"
)

// TokenAuth uses a preexisting token to authenticate to Cerberus
type TokenAuth struct {
	token   string
	headers http.Header
	baseURL *url.URL
}

// NewTokenAuth takes a Cerberus URL and valid token and returns a new TokenAuth.
// There is no checking done on whether or not the token is valid, so the function
// expects the a valid token. The URL and token can also be set using the CERBERUS_URL
// and CERBERUS_TOKEN environment variables. These will always take precedence over
// any arguments to the function
func NewTokenAuth(cerberusURL, token string) (*TokenAuth, error) {
	// Check for the environment variable if the user has set it
	if os.Getenv("CERBERUS_URL") != "" {
		cerberusURL = os.Getenv("CERBERUS_URL")
	}
	// Check for the environment variable for the token if the user has set it
	if os.Getenv("CERBERUS_TOKEN") != "" {
		token = os.Getenv("CERBERUS_TOKEN")
	}
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
	headers.Set("X-Vault-Token", token)
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
	t.headers.Set("X-Vault-Token", r.Data.ClientToken.ClientToken)
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
	t.headers.Del("X-Vault-Token")
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
