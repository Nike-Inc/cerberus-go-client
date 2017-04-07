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

// DoRequest is used to perform an HTTP request with the given method and path
// This method is what is called by other parts of the client and is exposed for advanced usage
func (c *Client) DoRequest(method, path string, data interface{}) (*http.Response, error) {
	// Get a copy of the base URL and add the path
	var baseURL = *c.CerberusURL
	baseURL.Path = path
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
