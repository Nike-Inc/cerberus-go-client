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

package cerberus

import (
	"fmt"
	"net/http"

	"github.com/Nike-Inc/cerberus-go-client/v3/api"
)

// Role is a subclient for accessing the roles endpoint
type Role struct {
	c *Client
}

var roleBasePath = "/v1/role"

// List returns a list of roles that can be granted
func (r *Role) List() ([]*api.Role, error) {
	resp, err := r.c.DoRequest(http.MethodGet, roleBasePath, map[string]string{}, nil)
	if resp != nil {
		defer resp.Body.Close()
	}
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
