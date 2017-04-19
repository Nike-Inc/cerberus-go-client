package cerberus

import (
	"fmt"
	"net/http"

	"github.nike.com/ngp/cerberus-client-go/api"
)

// Category is a subclient for accessing the category endpoint
type Category struct {
	c *Client
}

var categoryBasePath = "/v1/category"

// List returns a list of roles that can be granted
func (r *Category) List() ([]*api.Category, error) {
	resp, err := r.c.DoRequest(http.MethodGet, categoryBasePath, map[string]string{}, nil)
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
