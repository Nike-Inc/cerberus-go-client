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

package cerberus

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/Nike-Inc/cerberus-go-client/api"
	"github.com/Nike-Inc/cerberus-go-client/auth"
	vault "github.com/hashicorp/vault/api"
)

// Client is the main client for interacting with Cerberus
type Client struct {
	Authentication auth.Auth
	CerberusURL    *url.URL
	vaultClient    *vault.Client
	httpClient     *http.Client
}

// NewClient creates a new Client given an Authentication method.
// This method expects a file (which can be nil) as a source for a OTP used for MFA against Cerberus (if needed).
// If it is a file, it expect the token and a new line.
func NewClient(authMethod auth.Auth, otpFile *os.File) (*Client, error) {
	// Get the token and authenticate
	token, loginErr := authMethod.GetToken(otpFile)
	if loginErr != nil {
		return nil, loginErr
	}
	// Setup the vault client
	vaultConfig := vault.DefaultConfig()
	vaultConfig.Address = authMethod.GetURL().String()
	vclient, clientErr := vault.NewClient(vaultConfig)
	if clientErr != nil {
		return nil, fmt.Errorf("Error while setting up vault client: %v", clientErr)
	}
	// Used the returned token to set it as the token for this client as well
	vclient.SetToken(token)
	return &Client{
		Authentication: authMethod,
		CerberusURL:    authMethod.GetURL(),
		vaultClient:    vclient,
		httpClient:     &http.Client{},
	}, nil
}

// SDB returns the SDB client
func (c *Client) SDB() *SDB {
	return &SDB{
		c: c,
	}
}

// Secret returns the Secret client
func (c *Client) Secret() *Secret {
	return &Secret{
		v: c.vaultClient.Logical(),
	}
}

// Role returns the Role client
func (c *Client) Role() *Role {
	return &Role{
		c: c,
	}
}

// Category returns the Category client
func (c *Client) Category() *Category {
	return &Category{
		c: c,
	}
}

// Metadata returns the Metadata client
func (c *Client) Metadata() *Metadata {
	return &Metadata{
		c: c,
	}
}

// SecureFile returns the SecureFile client
func (c *Client) SecureFile() *SecureFile {
	return &SecureFile{
		c: c,
	}
}

// ErrorBodyNotReturned is an error indicating that the server did not return error details (in case of a non-successful status).
// This likely means that there is some sort of server error that is occurring
var ErrorBodyNotReturned = fmt.Errorf("No error body returned from server")

// DoRequestWithBody executes a request with provided body
func (c *Client) DoRequestWithBody(method, path string, params map[string]string, contentType string, body io.Reader) (*http.Response, error) {
	// Get a copy of the base URL and add the path
	var baseURL = *c.CerberusURL
	baseURL.Path = path
	p := baseURL.Query()
	// Add the params in to the request
	for k, v := range params {
		p.Add(k, v)
	}
	baseURL.RawQuery = p.Encode()
	var req *http.Request
	var err error

	req, err = http.NewRequest(method, baseURL.String(), body)
	if err != nil {
		return nil, err
	}
	headers, headerErr := c.Authentication.GetHeaders()
	if headerErr != nil {
		return nil, headerErr
	}
	req.Header = headers

	// Add content type if present
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	resp, respErr := c.httpClient.Do(req)
	if respErr != nil {
		return nil, respErr
	}
	// Cerberus uses a refresh token header. If that header is sent with a value of "true,"
	// refresh the token before returning
	if resp.Header.Get("X-Refresh-Token") == "true" {
		if err := c.Authentication.Refresh(); err != nil {
			// logging here
		}
		tok, err := c.Authentication.GetToken(nil)
		if err != nil {
			return nil, err
		}
		// Used the returned token to set it as the token for this client as well
		c.vaultClient.SetToken(tok)
	}
	return resp, nil
}

// DoRequest is used to perform an HTTP request with the given method and path
// This method is what is called by other parts of the client and is exposed for advanced usage
func (c *Client) DoRequest(method, path string, params map[string]string, data interface{}) (*http.Response, error) {
	var body io.ReadWriter
	var contentType string

	if data != nil {
		body = &bytes.Buffer{}
		contentType = "application/json"
		err := json.NewEncoder(body).Encode(data)
		if err != nil {
			return nil, err
		}
	}

	return c.DoRequestWithBody(method, path, params, contentType, body)
}

// parseResponse marshals the given body into the given interface. It should be used just like
// json.Marshal in that you pass a pointer to the function.
func parseResponse(r io.Reader, parseTo interface{}) error {
	// Decode the body into the provided interface
	return json.NewDecoder(r).Decode(parseTo)
}

// handleAPIError is a helper for parsing an error response body from the API.
// If the body doesn't have an error, it will return ErrorBodyNotReturned to indicate that there was no error body sent (probably means there was a server error)
func handleAPIError(r io.Reader) error {
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
