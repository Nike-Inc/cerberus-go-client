package auth

import (
	"fmt"
	"net/http"
	"os"
)

// ErrorUnauthenticated is used when a user tries to Refresh or Logout without already being authenticated
var ErrorUnauthenticated = fmt.Errorf("Unable to complete request: Not Authenticated")

// ErrorUnauthorized is returned when the request fails because of invalid credentials
var ErrorUnauthorized = fmt.Errorf("Invalid credentials given")

// The Auth interface describes the methods that all authentication providers must satisfy
type Auth interface {
	// GetToken should either return an existing token or perform all authentication steps
	// necessary to get a new token. It takes a file object as an argument as a place to
	// read an OTP for MFA flow
	GetToken(*os.File) (string, error)
	// IsAuthenticated should return whether or not there is a valid token. A valid token
	// is one that exists and is not expired
	IsAuthenticated() bool
	// Refresh uses the current valid token to retrieve a new one
	Refresh() error
	// Logout revokes the current token
	Logout() error
	// GetHeaders is a helper for any client using the authentication strategy.
	// It returns a basic set of headers asking for a JSON response and has
	// the authorization header set with the proper token
	GetHeaders() (http.Header, error)
}
