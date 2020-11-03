/*
Copyright 2019 Nike Inc.

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
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Nike-Inc/cerberus-go-client/v3/api"
	"github.com/Nike-Inc/cerberus-go-client/v3/utils"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/defaults"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/signer/v4"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"os"
	"time"
)

// STSAuth uses AWS V4 signing authenticate to Cerberus.
type STSAuth struct {
	token   string
	region  string
	expiry  time.Time
	baseURL *url.URL
	headers http.Header
}

// NewSTSAuth returns an STSAuth given a valid URL and region.
// Valid AWS credentials configured either by environment or through a credentials
// config file are also required.
func NewSTSAuth(cerberusURL, region string) (*STSAuth, error) {
	if len(region) == 0 {
		return nil, fmt.Errorf("Region cannot be empty")
	}
	if len(cerberusURL) == 0 {
		return nil, fmt.Errorf("Cerberus URL cannot be empty")
	}
	parsedURL, err := utils.ValidateURL(cerberusURL)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, fmt.Errorf("Unable to create AWS session: %s", err)
	}
	return &STSAuth{
		region:  region,
		baseURL: parsedURL,
		headers: http.Header{
			"Content-Type": []string{"application/json"},
		},
	}, nil
}

// GetToken returns a token if it already exists and is not expired. Otherwise,
// it authenticates using the provided URL and region and then returns the token.
func (a *STSAuth) GetToken(*os.File) (string, error) {
	if a.IsAuthenticated() {
		return a.token, nil
	}
	err := a.authenticate()
	return a.token, err
}

// GetExpiry returns the expiry time of the token if it already exists. Otherwise,
// it returns a zero-valued time.Time struct and an error.
func (a *STSAuth) GetExpiry() (time.Time, error) {
	if len(a.token) > 0 {
		return a.expiry, nil
	}
	return time.Time{}, fmt.Errorf("Expiry time not set.")
}

func (a *STSAuth) authenticate() error {
	builtURL := *a.baseURL
	builtURL.Path = "v2/auth/sts-identity"
	body := bytes.NewReader([]byte("Action=GetCallerIdentity&Version=2011-06-15"))

	request, err := http.NewRequest("POST", builtURL.String(), body)
	if err != nil {
		return fmt.Errorf("Problem while creating request to Cerberus: %v", err)
	}

	headers, err := a.sign()
	if err != nil {
		return fmt.Errorf("Problem signing request to Cerberus: %v", err)
	}
	for k, v := range headers {
		request.Header.Set(k, v[0])
	}

	client := http.Client{Timeout: 10 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("Problem while performing request to Cerberus: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusUnauthorized || response.StatusCode == http.StatusForbidden {
		return fmt.Errorf("Invalid credentials given. Verify that the role you are currently using is valid " +
			"with the AWS CLI ($ aws sts get-caller-identity) or with gimme-aws-creds.")
	}
	if response.StatusCode != http.StatusOK {
		apiErr := utils.ParseAPIError(response.Body)
		return fmt.Errorf("Error while trying to authenticate. Got HTTP response code %d\n%v", response.StatusCode, apiErr)
	}

	decoder := json.NewDecoder(response.Body)
	authResponse := &api.IAMAuthResponse{}
	dErr := decoder.Decode(authResponse)
	if dErr != nil {
		return fmt.Errorf("Error while trying to parse response from Cerberus: %v", err)
	}

	metadata := authResponse.Metadata
	identity := "unknown"
	if iam_principal, found := metadata["iam_principal_arn"]; found {
		identity = iam_principal
	} else if username, found := metadata["username"]; found {
		identity = username
	}
	log.Info(fmt.Sprintf("Successfully authenticated with Cerberus as %v\n", identity))

	a.token = authResponse.Token
	a.headers.Set("X-Cerberus-Token", authResponse.Token)
	a.expiry = time.Now().Add((time.Duration(authResponse.Duration) * time.Second) - expiryDelta)
	return nil
}

// IsAuthenticated returns whether or not the current token is set and is not expired.
func (a *STSAuth) IsAuthenticated() bool {
	return len(a.token) > 0 && time.Now().Before(a.expiry)
}

// Refresh refreshes the current token by reauthenticating against the API.
func (a *STSAuth) Refresh() error {
	if !a.IsAuthenticated() {
		return api.ErrorUnauthenticated
	}
	// A note on why we are just reauthenticating: You can refresh an AWS token,
	// but there is a limit (24) to the number of refreshes and the API requests
	// that you refresh your token on every SDB creation. When doing this in an
	// automation context, you could surpass this limit. You could not refresh
	// the token, but it can get you in to a state where you can't perform some
	// operations. This is less than ideal but better than having an arbitary
	// bound on the number of refreshes and having to track how many have been
	// done.
	return a.authenticate()
}

// Logout deauthorizes the current valid token. This will return an error if the token
// is expired or non-existent.
func (a *STSAuth) Logout() error {
	if !a.IsAuthenticated() {
		return api.ErrorUnauthenticated
	}
	// Use a copy of the base URL
	if err := Logout(*a.baseURL, a.headers); err != nil {
		return err
	}
	// Reset the token and header
	a.token = ""
	a.headers.Del("X-Cerberus-Token")
	return nil
}

// GetHeaders returns the headers needed to authenticate against Cerberus. This will
// return an error if the token is expired or non-existent.
func (a *STSAuth) GetHeaders() (http.Header, error) {
	if !a.IsAuthenticated() {
		return nil, api.ErrorUnauthenticated
	}
	return a.headers, nil
}

// GetURL returns the configured Cerberus URL.
func (a *STSAuth) GetURL() *url.URL {
	return a.baseURL
}

// credentials obtains default AWS credentials.
func creds() *credentials.Credentials {
	creds := defaults.Get().Config.Credentials
	return creds
}

// signer returns a V4 signer for signing a request.
func signer() (*v4.Signer, error) {
	creds := creds()
	_, err := creds.Get()
	if err != nil {
		return nil, fmt.Errorf("Credentials are required and cannot be found: %v", err)
	}
	signer := v4.NewSigner(creds)
	return signer, nil
}

// request creates an STS Auth request.
func (a *STSAuth) request() (*http.Request, error) {

	var chinaRegions = make(map[string]struct{})
	chinaRegions["cn-north-1"] = struct{}{}
	chinaRegions["cn-northwest-1"] = struct{}{}

	_, err := endpoints.DefaultResolver().EndpointFor("sts", a.region, endpoints.StrictMatchingOption)
	if err != nil {
		return nil, fmt.Errorf("Endpoint could not be created. "+
			"Confirm that region, %v, is a valid AWS region : %v", a.region, err)
	}
	method := "POST"
	url := "https://sts." + a.region + ".amazonaws.com"
	if _, ok := chinaRegions[a.region]; ok {
		url += ".cn"
	}
	request, _ := http.NewRequest(method, url, nil)
	return request, nil
}

// sign signs a AWS v4 request and returns the signed headers.
func (a *STSAuth) sign() (http.Header, error) {
	signer, signErr := signer()
	if signErr != nil {
		return nil, signErr
	}
	request, reqErr := a.request()
	if reqErr != nil {
		return nil, reqErr
	}
	service := "sts"
	body := bytes.NewReader([]byte("Action=GetCallerIdentity&Version=2011-06-15"))

	_, signerErr := signer.Sign(request, body, service, a.region, time.Now())
	if signerErr != nil {
		return nil, signerErr
	}
	return request.Header, nil
}
