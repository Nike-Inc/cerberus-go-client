package cerberus

import (
	"fmt"
	"net/http"

	"github.com/Nike-Inc/cerberus-go-client/api"
)

// Role is a subclient for accessing the roles endpoint
type Role struct {
	c *Client
}

var roleBasePath = "/v1/role"

// List returns a list of roles that can be granted
func (r *Role) List() ([]*api.Role, error) {
	resp, err := r.c.DoRequest(http.MethodGet, roleBasePath, map[string]string{}, nil)
	if err != nil {
		return nil, fmt.Errorf("Error while trying to get roles: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Error while trying to GET roles. Got HTTP status code %d", resp.StatusCode)
	}
	var roleList = []*api.Role{}
	err = parseResponse(resp.Body, &roleList)
	if err != nil {
		return nil, err
	}
	return roleList, nil
}
