package cerberus

import (
	"fmt"
	"net/http"
)

//"github.nike.com/ngp/cerberus-client-go/auth"

type CerberusClient struct {
	authenticated bool
	baseHeaders   http.Header
}

func NewCerberusClientFromPassword(cerberusURL, username, password string) (*CerberusClient, error) {
	return nil, nil
}

func NewCerberusClientFromAWS(cerberusURL, roleARN, region string) (*CerberusClient, error) {
	return nil, fmt.Errorf("Unimplemented")
}
