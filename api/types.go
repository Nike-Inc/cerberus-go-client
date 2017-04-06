// Package api contains the Cerberus API object definitions
// This is not a full implementation of every object right now
// and only defines the needed objects for the client to function.
// See https://github.com/Nike-Inc/cerberus-management-service/blob/master/API.md
// for full documentation
package api

// AuthStatus is the status of a UserAuthResponse
type AuthStatus string

var (
	// AuthUserSuccess indicates that the username/password login was successful
	AuthUserSuccess AuthStatus = "success"
	// AuthUserNeedsMFA indicates that the username/password login was successful
	// but an MFA token is required
	AuthUserNeedsMFA AuthStatus = "mfa_req"
)

// UserAuthResponse represents the response from the /v2/auth/user
type UserAuthResponse struct {
	Status AuthStatus
	Data   UserAuthData
}

// UserAuthData contains user dat for the authentication request or for MFA verification
type UserAuthData struct {
	ClientToken UserClientToken `json:"client_token"`
	UserID      string          `json:"user_id"`
	Username    string
	StateToken  string `json:"state_token"`
	Devices     []MFADevice
}

// UserClientToken represents the authentication token returned from the API
type UserClientToken struct {
	ClientToken string `json:"client_token"`
	Policies    []string
	Metadata    UserMetadata
	Duration    int `json:"lease_duration"`
	Renewable   bool
}

// MFADevice represents a user method for providing a token
type MFADevice struct {
	ID   string
	Name string
}

// UserMetadata represents the user data to which a token belongs
type UserMetadata struct {
	Username string
	IsAdmin  string `json:"is_admin"`
	Groups   string
}
