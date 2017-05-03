package cerberus

import (
	"fmt"
	"net/http"

	"github.com/Nike-Inc/cerberus-go-client/api"
)

// Metadata is a subclient for accessing the metadata endpoint
type Metadata struct {
	c *Client
}

// MetadataOpts is used for passing pagination values to the list function
type MetadataOpts struct {
	Limit  uint
	Offset uint
}

var metadataBasePath = "/v1/metadata"

// List returns a MetadataResponse which is a wrapper containing pagination data and an array of metadata objects
func (m *Metadata) List(opts MetadataOpts) (*api.MetadataResponse, error) {
	// Set the limit opt to default if it isn't set
	if opts.Limit == 0 {
		opts.Limit = 100
	}
	// Put the options into the params
	var params = map[string]string{}
	params["limit"] = fmt.Sprintf("%d", opts.Limit)
	params["offset"] = fmt.Sprintf("%d", opts.Offset)
	resp, err := m.c.DoRequest(http.MethodGet, metadataBasePath, params, nil)
	if err != nil {
		return nil, fmt.Errorf("Error while trying to get roles: %v", err)
	}
	// Check if it is a bad request (improperly set params)
	if resp.StatusCode == http.StatusBadRequest {
		// Return the API error to the user
		return nil, handleAPIError(resp.Body)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Error while trying to GET metadata. Got HTTP status code %d", resp.StatusCode)
	}
	var metadataResp = &api.MetadataResponse{}
	err = parseResponse(resp.Body, metadataResp)
	if err != nil {
		return nil, err
	}
	return metadataResp, nil
}
