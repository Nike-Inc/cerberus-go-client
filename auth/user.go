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
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/Nike-Inc/cerberus-go-client/api"
	"github.com/Nike-Inc/cerberus-go-client/utils"
)

// UserAuth uses username and password authentication to authenticate against Cerberus
type UserAuth struct {
	username string
	password string
	baseURL  *url.URL
	token    string
	expiry   time.Time
	headers  http.Header
	client   *http.Client
}

// NewUserAuth returns a new UserAuth object given a valid Cerberus URL, username, and password
func NewUserAuth(cerberusURL, username, password string) (*UserAuth, error) {
	// Check for the environment variable if the user has set it
	if os.Getenv("CERBERUS_URL") != "" {
		cerberusURL = os.Getenv("CERBERUS_URL")
	}
	// Make sure there isn't a blank username, password, or URL
	if len(username) == 0 {
		return nil, fmt.Errorf("Username cannot be empty")
	}
	if len(password) == 0 {
		return nil, fmt.Errorf("Password cannot be empty")
	}
	if len(cerberusURL) == 0 {
		return nil, fmt.Errorf("Cerberus URL cannot be empty")
	}
	parsedURL, err := utils.ValidateURL(cerberusURL)
	if err != nil {
		return nil, err
	}
	return &UserAuth{
		username: username,
		password: password,
		baseURL:  parsedURL,
		headers: http.Header{
			"Content-Type":      []string{"application/json"},
			"X-Cerberus-Client": []string{api.ClientHeader},
		},
		client: &http.Client{},
	}, nil
}

// GetToken returns an existing token or performs all authentication steps
// necessary to get a new token. This should be called to authenticate the
// client once it has been setup
func (u *UserAuth) GetToken(f *os.File) (string, error) {
	if u.IsAuthenticated() {
		return u.token, nil
	}
	// Try to log in
	if err := u.authenticate(f); err != nil {
		return "", err
	}
	return u.token, nil
}

// GetExpiry returns the expiry time of the token if it already exists. Otherwise,
// it returns a zero-valued time.Time struct and an error.
func (u *UserAuth) GetExpiry() (time.Time, error) {
	if len(u.token) > 0 {
		return u.expiry, nil
	}
	return time.Time{}, fmt.Errorf("Expiry time not set")
}

// GetURL returns the URL used for Cerberus
func (u *UserAuth) GetURL() *url.URL {
	return u.baseURL
}

// IsAuthenticated returns whether or not there is a valid token. A valid token
// is one that exists and is not expired
func (u *UserAuth) IsAuthenticated() bool {
	return len(u.token) > 0 && time.Now().Before(u.expiry)
}

// Refresh uses the current valid token to retrieve a new one. Returns
// ErrorUnauthenticated if not already authenticated
func (u *UserAuth) Refresh() error {
	if !u.IsAuthenticated() {
		return api.ErrorUnauthenticated
	}
	// Pass a copy of the base URL
	r, err := Refresh(*u.baseURL, u.headers)
	if err != nil {
		return err
	}
	u.setToken(r.Data.ClientToken.ClientToken, r.Data.ClientToken.Duration)
	return nil
}

// Logout revokes the current token. Returns ErrorUnauthenticated if
// not already authenticated
func (u *UserAuth) Logout() error {
	if !u.IsAuthenticated() {
		return api.ErrorUnauthenticated
	}
	// Use a copy of the base URL
	if err := Logout(*u.baseURL, u.headers); err != nil {
		return err
	}
	// Reset the token and header
	u.token = ""
	u.headers.Del("X-Cerberus-Token")
	return nil
}

// GetHeaders is a helper for any client using the authentication strategy.
// It returns a basic set of headers asking for a JSON response and has
// the authorization header set with the proper token
func (u *UserAuth) GetHeaders() (http.Header, error) {
	if !u.IsAuthenticated() {
		return nil, api.ErrorUnauthenticated
	}
	return u.headers, nil
}

func (u *UserAuth) authenticate(f *os.File) error {
	encodedCreds := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", u.username, u.password)))
	headers := http.Header{
		"Authorization":     []string{fmt.Sprintf("Basic %s", encodedCreds)},
		"X-Cerberus-Client": []string{api.ClientHeader},
	}
	// Make a copy of the base URL
	builtURL := *u.baseURL
	builtURL.Path = "/v2/auth/user"
	req, err := http.NewRequest("GET", builtURL.String(), nil)
	if err != nil {
		return err
	}
	req.Header = headers
	resp, err := u.client.Do(req)
	if err != nil {
		return fmt.Errorf("Problem while performing request to Cerberus: %v", err)
	}
	r, checkErr := utils.CheckAndParse(resp)
	if checkErr != nil {
		return checkErr
	}
	// Check for MFA
	if r.Status == api.AuthUserNeedsMFA {
		// If MFA is enabled, there should always be at least one device
		// TODO: This ain't pretty because it only works for one device. See comment in doMFA as well
		return u.doMFA(r.Data.StateToken, r.Data.Devices[0].ID, f)
	}
	u.setToken(r.Data.ClientToken.ClientToken, r.Data.ClientToken.Duration)
	return nil
}

// doMFA is the handler for MFA and reads a OTP token from a file. If file is nil, os.Stdin is used
func (u *UserAuth) doMFA(stateToken, deviceID string, readFrom *os.File) error {
	// TODO: There has got to be a smarter way to do this. This is copied from the python client logic
	var body = map[string]string{
		"device_id":   deviceID,
		"state_token": stateToken,
	}
	var source *os.File
	// Set the source of the input
	if readFrom == nil {
		source = os.Stdin
	} else {
		source = readFrom
	}
	// Capture the OTP from the user
	reader := bufio.NewReader(source)
	// Only print a prompt if the source is stdin
	if source == os.Stdin {
		fmt.Print("Enter token from device: ")
	}
	token, _ := reader.ReadString('\n')
	// Clean it up and put it in the body
	body["otp_token"] = strings.TrimSpace(token)
	// Make a copy of the base URL
	builtURL := *u.baseURL
	builtURL.Path = "/v2/auth/mfa_check"
	// Put the body into a buffer
	data := &bytes.Buffer{}
	if err := json.NewEncoder(data).Encode(body); err != nil {
		return fmt.Errorf("Error while trying to encode MFA response: %v", err)
	}
	resp, err := http.Post(builtURL.String(), "application/json", data)
	if err != nil {
		return fmt.Errorf("Problem while performing request to Cerberus: %v", err)
	}
	r, checkErr := utils.CheckAndParse(resp)
	if checkErr != nil {
		return checkErr
	}
	u.setToken(r.Data.ClientToken.ClientToken, r.Data.ClientToken.Duration)
	return nil
}

// setToken is a helper method so that both the traditional and MFA user auth methods can set the token
// without repeating any logic
func (u *UserAuth) setToken(token string, duration int) {
	u.token = token
	// Set the auth header up to make things easier
	u.headers.Set("X-Cerberus-Token", token)
	u.expiry = time.Now().Add((time.Duration(duration) * time.Second) - expiryDelta)
}
