package cerberus

import (
	"fmt"
	"net/http"

	"github.nike.com/ngp/cerberus-client-go/api"
)

// ErrorSafeDepositBoxNotFound is returned when a specified deposit box is not found
var ErrorSafeDepositBoxNotFound = fmt.Errorf("Unable to find Safe Deposit Box")

var basePath = "/v1/safe-deposit-box"

// SDB is a client for managing and reading SafeDepositBox objects
type SDB struct {
	// a pointer to its parent client
	c *Client
}

// GetByName is a helper method that takes a SDB name and attempts
// to locate that box in a list of SDBs the client has access to
func (s *SDB) GetByName(name string) (*api.SafeDepositBox, error) {
	if len(name) == 0 {
		return nil, ErrorSafeDepositBoxNotFound
	}
	allSDB, err := s.List()
	if err != nil {
		return nil, err
	}
	for _, v := range allSDB {
		if v.Name == name {
			return v, nil
		}
	}
	// If we didn't find it in the list, return error that it wasn't found
	return nil, ErrorSafeDepositBoxNotFound
}

// Get returns a single SDB given an ID. Returns ErrorSafeDepositBoxNotFound
// if the ID does not exist
func (s *SDB) Get(id string) (*api.SafeDepositBox, error) {
	if len(id) == 0 {
		return nil, ErrorSafeDepositBoxNotFound
	}
	returnedSDB := &api.SafeDepositBox{}
	resp, err := s.c.DoRequest("GET", basePath+"/"+id, nil)
	if err != nil {
		return nil, fmt.Errorf("Error while trying to get SDB: %v", err)
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrorSafeDepositBoxNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Error while trying to GET SDB. Got HTTP status code %d", resp.StatusCode)
	}
	err = parseResponse(resp.Body, returnedSDB)
	if err != nil {
		return nil, err
	}
	return returnedSDB, nil
}

// List returns a list of all SDBs the authenticated user is allowed to see
func (s *SDB) List() ([]*api.SafeDepositBox, error) {
	sdbList := []*api.SafeDepositBox{}
	resp, err := s.c.DoRequest("GET", basePath, nil)
	if err != nil {
		return nil, fmt.Errorf("Error while trying to lidy SDB: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Error while trying to GET SDB list. Got HTTP status code %d", resp.StatusCode)
	}
	err = parseResponse(resp.Body, &sdbList)
	if err != nil {
		return nil, err
	}
	return sdbList, nil
}

// Create creates a new Safe Deposit Box and returns the newly created object
func (s *SDB) Create(newSDB api.SafeDepositBox) (*api.SafeDepositBox, error) {
	return nil, fmt.Errorf("Unimplemented")
}

// Update updates an existing Safe Deposit Box. Any fields that are not null in the passed object
// will overwrite any fields on the current object
func (s *SDB) Update(id string, updatedSDB api.SafeDepositBox) (*api.SafeDepositBox, error) {
	return nil, fmt.Errorf("Unimplemented")
}

// Delete deletes the Safe Deposit Box with the given ID
func (s *SDB) Delete(id string) error {
	return fmt.Errorf("Unimplemented")
}
