package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/Nike-Inc/cerberus-go-client/api"
	"github.com/Nike-Inc/cerberus-go-client/utils"
)

// AWSAuth uses AWS roles and authentication to authenticate to Cerberus
type AWSAuth struct {
	token   string
	region  string
	roleARN string
	expiry  time.Time
	baseURL *url.URL
	headers http.Header
}

type awsAuthBody struct {
	PrincipalArn string `json:"iam_principal_arn"`
	Region       string `json:"region"`
}

// NewAWSAuth returns an AWSAuth given a valid URL, ARN, and region. If the CERBERUS_URL
// environment variable is set, it will be used over anything passed to this function
func NewAWSAuth(cerberusURL, roleARN, region string) (*AWSAuth, error) {
	// Check for the environment variable if the user has set it
	if os.Getenv("CERBERUS_URL") != "" {
		cerberusURL = os.Getenv("CERBERUS_URL")
	}
	if len(roleARN) == 0 {
		return nil, fmt.Errorf("Role ARN should not be empty")
	}
	if len(region) == 0 {
		return nil, fmt.Errorf("Region should not be nil")
	}
	if len(cerberusURL) == 0 {
		return nil, fmt.Errorf("Cerberus URL cannot be empty")
	}
	parsedURL, err := utils.ValidateURL(cerberusURL)
	if err != nil {
		return nil, err
	}
	return &AWSAuth{
		region:  region,
		roleARN: roleARN,
		baseURL: parsedURL,
		headers: http.Header{},
	}, nil
}

// GetURL returns the configured Cerberus URL
func (a *AWSAuth) GetURL() *url.URL {
	return a.baseURL
}

// GetToken returns a token if it already exists and is not expired. Otherwise,
// it authenticates using the provided ARN and region and then returns the token.
// If there are any errors during authentication,
func (a *AWSAuth) GetToken(f *os.File) (string, error) {
	if a.IsAuthenticated() {
		return a.token, nil
	}
	// Make a copy of the base URL
	builtURL := *a.baseURL
	builtURL.Path = "/v2/auth/iam-principal"
	// Encode the body to send in the request if one was given
	body := &bytes.Buffer{}
	err := json.NewEncoder(body).Encode(awsAuthBody{
		PrincipalArn: a.roleARN,
		Region:       a.region,
	})
	if err != nil {
		return "", err
	}
	resp, err := http.Post(builtURL.String(), "application/json", body)
	if err != nil {
		return "", fmt.Errorf("Problem while performing request to Cerberus: %v", err)
	}
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return "", api.ErrorUnauthorized
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Error while trying to authenticate. Got HTTP response code %d", resp.StatusCode)
	}
	decoder := json.NewDecoder(resp.Body)
	r := &api.IAMAuthResponse{}
	dErr := decoder.Decode(r)
	if dErr != nil {
		return "", fmt.Errorf("Error while trying to parse response from Cerberus: %v", err)
	}
	a.token = r.Token
	// Set the auth header up to make things easier
	a.headers.Set("X-Vault-Token", r.Token)
	a.expiry = time.Now().Add(time.Duration(r.Duration) * time.Second)
	return a.token, nil
}

// IsAuthenticated returns whether or not the current token is set and is not expired
func (a *AWSAuth) IsAuthenticated() bool {
	return len(a.token) > 0 && time.Now().Before(a.expiry)
}

// Refresh refreshes the current token
func (a *AWSAuth) Refresh() error {
	if !a.IsAuthenticated() {
		return api.ErrorUnauthenticated
	}
	r, err := Refresh(*a.baseURL, a.headers)
	if err != nil {
		return err
	}
	a.token = r.Data.ClientToken.ClientToken
	a.expiry = time.Now().Add(time.Duration(r.Data.ClientToken.Duration) * time.Second)
	a.headers.Set("X-Vault-Token", r.Data.ClientToken.ClientToken)
	return nil
}

// Logout deauthorizes the current valid token. This will return an error if the token
// is expired or non-existent
func (a *AWSAuth) Logout() error {
	if !a.IsAuthenticated() {
		return api.ErrorUnauthenticated
	}
	// Use a copy of the base URL
	if err := Logout(*a.baseURL, a.headers); err != nil {
		return err
	}
	// Reset the token and header
	a.token = ""
	a.headers.Del("X-Vault-Token")
	return nil
}

// GetHeaders returns the headers needed to authenticate against Cerberus. This will
// return an error if the token is expired or non-existent
func (a *AWSAuth) GetHeaders() (http.Header, error) {
	if !a.IsAuthenticated() {
		return nil, api.ErrorUnauthenticated
	}
	return a.headers, nil
}
