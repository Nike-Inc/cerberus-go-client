package cerberus

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	vault "github.com/hashicorp/vault/api"
	"github.nike.com/ngp/cerberus-client-go/api"
	"github.nike.com/ngp/cerberus-client-go/auth"
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

// ErrorBodyNotReturned is an error indicating that the server did not return error details (in case of a non-successful status).
// This likely means that there is some sort of server error that is occurring
var ErrorBodyNotReturned = fmt.Errorf("No error body returned from server")

// DoRequest is used to perform an HTTP request with the given method and path
// This method is what is called by other parts of the client and is exposed for advanced usage
func (c *Client) DoRequest(method, path string, params map[string]string, data interface{}) (*http.Response, error) {
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
	if data == nil {
		req, err = http.NewRequest(method, baseURL.String(), nil)
	} else {
		// Encode the body to send in the request if one was given
		body := &bytes.Buffer{}
		err := json.NewEncoder(body).Encode(data)
		if err != nil {
			return nil, err
		}
		req, err = http.NewRequest(method, baseURL.String(), body)
	}

	if err != nil {
		return nil, err
	}
	headers, headerErr := c.Authentication.GetHeaders()
	if headerErr != nil {
		return nil, headerErr
	}
	req.Header = headers
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
	}
	return resp, nil
}

// parseResponse marshals the given body into the given interface. It should be used just like
// json.Marshal in that you pass a pointer to the function.
func parseResponse(r io.Reader, parseTo interface{}) error {
	// Decode the body into the provided interface
	if err := json.NewDecoder(r).Decode(parseTo); err != nil {
		return err
	}
	return nil
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
