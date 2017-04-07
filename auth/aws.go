package auth

import (
	"net/http"
	"net/url"
	"os"
	"time"
)

// AWSAuth uses AWS roles and authentication to authenticate to Cerberus
type AWSAuth struct {
	token   string
	region  string
	roleARN string
	expiry  time.Time
	baseURL *url.URL
}

func NewAWSAuth(cerberusURL, roleARN, region string) (*AWSAuth, error) {
	// Check for the environment variable if the user has set it
	if os.Getenv("CERBERUS_URL") != "" {
		cerberusURL = os.Getenv("CERBERUS_URL")
	}
	return &AWSAuth{}, nil
}

func (a *AWSAuth) GetURL() *url.URL {
	return a.baseURL
}

func (a *AWSAuth) GetToken(f *os.File) (string, error) {
	return "", nil
}

func (a *AWSAuth) IsAuthenticated() bool {
	return len(a.token) > 0 && time.Now().Before(a.expiry)
}

func (a *AWSAuth) Refresh() error {
	return nil
}

func (a *AWSAuth) Logout() error {
	return nil
}

func (a *AWSAuth) GetHeaders() (http.Header, error) {
	return nil, nil
}
