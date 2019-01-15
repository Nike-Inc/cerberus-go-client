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

// Package api contains the Cerberus API object definitions
// This is not a full implementation of every object right now
// and only defines the needed objects for the client to function.
// See https://github.com/Nike-Inc/cerberus-management-service/blob/master/API.md
// for full documentation
package api

import (
	"fmt"
	"time"
)

// ClientHeader is the header version for all requests. It should be updated on version bumps
const ClientHeader = "CerberusGoClient/0.3.1"

// AuthStatus is the status of a UserAuthResponse
type AuthStatus string

var (
	// AuthUserSuccess indicates that the username/password login was successful
	AuthUserSuccess AuthStatus = "success"
	// AuthUserNeedsMFA indicates that the username/password login was successful
	// but an MFA token is required
	AuthUserNeedsMFA AuthStatus = "mfa_req"
)

// ErrorUnauthenticated is used when a user tries to Refresh or Logout without already being authenticated
var ErrorUnauthenticated = fmt.Errorf("Unable to complete request: Not Authenticated")

// ErrorUnauthorized is returned when the request fails because of invalid credentials
var ErrorUnauthorized = fmt.Errorf("Invalid credentials given")

// ErrorResponse represents an error response from the API
type ErrorResponse struct {
	ErrorID string `json:"error_id"`
	Errors  []ErrorDetail
}

// ErrorDetail is a specific error description for a given issue. There may be many of these returned with an ErrorResponse
type ErrorDetail struct {
	Code     int
	Message  string
	Metadata map[string]interface{} // Most of the time it is just a string => string. But the error definition states this as an "Object" in Java, so it could be anything
}

func (e ErrorResponse) Error() string {
	return fmt.Sprintf("Error from API. ID: %s, Details: %+v", e.ErrorID, e.Errors)
}

// IAMAuthResponse represents a response from the iam-principal authentication endpoint
type IAMAuthResponse struct {
	Token     string `json:"client_token"`
	Policies  []string
	Metadata  AWSMetadata
	Duration  int `json:"lease_duration"`
	Renewable bool
}

// AWSMetadata contains additional information about the ARN that was used to log in
type AWSMetadata struct {
	Region       string `json:"aws_region"`
	PrincipalARN string `json:"iam_principal_arn"`
	Username     string
	IsAdmin      string `json:"is_admin"` // This is returned as a string from the API
	Groups       string
}

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

// SafeDepositBox represents a safe deposit box API object
type SafeDepositBox struct {
	ID                      string                `json:"id,omitempty"`
	Name                    string                `json:"name,omitempty"`
	Path                    string                `json:"path,omitempty"`
	CategoryID              string                `json:"category_id,omitempty"`
	Description             string                `json:"description,omitempty"`
	Owner                   string                `json:"owner,omitempty"`
	UserGroupPermissions    []UserGroupPermission `json:"user_group_permissions,omitempty"`
	IAMPrincipalPermissions []IAMPrincipal        `json:"iam_principal_permissions,omitempty"`
}

// UserGroupPermission represents a user and group permission on an object
type UserGroupPermission struct {
	ID     string
	Name   string `json:"name"`
	RoleID string `json:"role_id"`
}

// IAMPrincipal represents an IAM permission on an object
type IAMPrincipal struct {
	ID              string
	IAMPrincipalARN string `json:"iam_principal_arn"`
	RoleID          string `json:"role_id"`
}

// Role represents a role that can be assigned to a safe deposit box
type Role struct {
	ID            string
	Name          string
	Created       time.Time `json:"created_ts"`
	LastUpdated   time.Time `json:"last_updated_ts"`
	CreatedBy     string    `json:"created_by"`
	LastUpdatedBy string    `json:"last_updated_by"`
}

// Category represents a category that can be assigned to a safe deposit box
type Category struct {
	ID            string
	DisplayName   string `json:"display_name"`
	Path          string
	Created       time.Time `json:"created_ts"`
	LastUpdated   time.Time `json:"last_updated_ts"`
	CreatedBy     string    `json:"created_by"`
	LastUpdatedBy string    `json:"last_updated_by"`
}

// MetadataResponse is an object that wraps a list of SDBMetadata for convenience with pagination
type MetadataResponse struct {
	HasNext     bool `json:"has_next"`
	NextOffset  int  `json:"next_offset"`
	Limit       int
	Offset      int
	ResultCount int           `json:"sdb_count_in_result"`
	TotalCount  int           `json:"total_sdbcount"`
	Metadata    []SDBMetadata `json:"safe_deposit_box_metadata"`
}

// SDBMetadata represents the metadata of a specific SDB
type SDBMetadata struct {
	Name                 string
	Path                 string
	Category             string
	Owner                string
	Description          string
	Created              time.Time         `json:"created_ts"`
	CreatedBy            string            `json:"created_by"`
	LastUpdated          time.Time         `json:"last_updated_ts"`
	LastUpdatedBy        string            `json:"last_updated_by"`
	UserGroupPermissions map[string]string `json:"user_group_permissions"`
	IAMRolePermissions   map[string]string `json:"iam_role_permissions"`
}

// SecureFileSummary represents the metadata of a specific secure-file
type SecureFileSummary struct {
	Name                 string            `json:"name"`
	Path                 string            `json:"path"`
	Size                 int               `json:"size_in_bytes"`
	SDBID                string            `json:"sdbox_id"`
	Created              time.Time         `json:"created_ts"`
	CreatedBy            string            `json:"created_by"`
	LastUpdated          time.Time         `json:"last_updated_ts"`
	LastUpdatedBy        string            `json:"last_updated_by"`
	UserGroupPermissions map[string]string `json:"user_group_permissions"`
	IAMRolePermissions   map[string]string `json:"iam_role_permissions"`
}

// SecureFilesResponse is an object that wraps a list of SecureFileSummary for convenience with pagination
type SecureFilesResponse struct {
	HasNext    bool `json:"has_next"`
	NextOffset int  `json:"next_offset"`
	Limit      int
	Offset     int

	ResultCount int                 `json:"file_count_in_result"`
	TotalCount  int                 `json:"total_file_count"`
	Summaries   []SecureFileSummary `json:"secure_file_summaries"`
}
