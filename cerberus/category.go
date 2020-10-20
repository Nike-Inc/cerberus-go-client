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

// Category is a subclient for accessing the category endpoint
type Category struct {
	c *Client
}

var categoryBasePath = "/v1/category"

// List returns a list of roles that can be granted
func (r *Category) List() ([]*api.Category, error) {
	resp, err := r.c.DoRequest(http.MethodGet, categoryBasePath, map[string]string{}, nil)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, fmt.Errorf("Error while trying to get categories: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Error while trying to GET categories. Got HTTP status code %d", resp.StatusCode)
	}
	var categoryList = []*api.Category{}
	err = parseResponse(resp.Body, &categoryList)
	if err != nil {
		return nil, err
	}
	return categoryList, nil
}
