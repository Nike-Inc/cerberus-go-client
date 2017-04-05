package auth

import (
	"net/http"
	"os"
	"time"
)

// AWSAuth uses AWS roles and authentication to authenticate to Cerberus
type AWSAuth struct {
	token   string
	region  string
	roleARN string
	expiry  time.Time
}

func NewAWSAuth(cerberusURL, roleARN, region string) (*AWSAuth, error) {
	return &AWSAuth{}, nil
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
