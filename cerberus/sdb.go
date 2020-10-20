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
	"github.com/Nike-Inc/cerberus-go-client/v3/utils"
	"net/http"
	"strings"

	"github.com/Nike-Inc/cerberus-go-client/v3/api"
)

// ErrorSafeDepositBoxNotFound is returned when a specified deposit box is not found
var ErrorSafeDepositBoxNotFound = fmt.Errorf("Unable to find Safe Deposit Box")

var sdbBasePath = "/v2/safe-deposit-box"

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
	resp, err := s.c.DoRequest(http.MethodGet, sdbBasePath+"/"+id, map[string]string{}, nil)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		if resp != nil {
			if resp.StatusCode == http.StatusNotFound {
				return nil, ErrorSafeDepositBoxNotFound
			}
		}
		return nil, fmt.Errorf("Error while trying to get SDB: %v", err)
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
	resp, err := s.c.DoRequest(http.MethodGet, sdbBasePath, map[string]string{}, nil)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, fmt.Errorf("Error while trying to list SDB: %v", err)
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
func (s *SDB) Create(newSDB *api.SafeDepositBox) (*api.SafeDepositBox, error) {
	// Create the object we are returning
	createdSDB := &api.SafeDepositBox{}
	resp, err := s.c.DoRequest(http.MethodPost, sdbBasePath, map[string]string{}, newSDB)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		// Check if it is a bad request (improperly set params)
		if resp != nil && resp.StatusCode == http.StatusBadRequest {
			// Return the API error to the user
			return nil, utils.ParseAPIError(resp.Body)
		}
		return nil, fmt.Errorf("Error while creating SDB: %v", err)
	}
	// If it isn't a bad request, make sure it is a good request and return an error if it isn't
	if resp.StatusCode != http.StatusCreated {
		apiErr := utils.ParseAPIError(resp.Body)
		if apiErr == ErrorBodyNotReturned {
			return nil, fmt.Errorf("Error while creating SDB. Got HTTP status code %d. %v", resp.StatusCode, apiErr)
		}
		return nil, apiErr
	}
	// Parse the created object
	err = parseResponse(resp.Body, createdSDB)
	if err != nil {
		return nil, err
	}
	return createdSDB, nil
}

// Update updates an existing Safe Deposit Box. Any fields that are not null in the passed object
// will overwrite any fields on the current object
func (s *SDB) Update(id string, updatedSDB *api.SafeDepositBox) (*api.SafeDepositBox, error) {
	id = strings.TrimSpace(id)
	// Check to make sure the ID isn't empty
	if id == "" {
		return nil, ErrorSafeDepositBoxNotFound
	}
	returnedSDB := &api.SafeDepositBox{}
	resp, err := s.c.DoRequest(http.MethodPut, sdbBasePath+"/"+id, map[string]string{}, updatedSDB)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		// Check if it is a bad request (improperly set params)
		if resp != nil {
			if resp.StatusCode == http.StatusNotFound {
				return nil, ErrorSafeDepositBoxNotFound
			}
			if resp.StatusCode == http.StatusBadRequest {
				// Return the API error to the user
				return nil, utils.ParseAPIError(resp.Body)
			}
		}
		return nil, fmt.Errorf("Error while updating SDB: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		apiErr := utils.ParseAPIError(resp.Body)
		if apiErr == ErrorBodyNotReturned {
			return nil, fmt.Errorf("Error while updating SDB. Got HTTP status code %d. %v", resp.StatusCode, apiErr)
		}
		return nil, apiErr
	}
	// Parse the updated object
	err = parseResponse(resp.Body, returnedSDB)
	if err != nil {
		return nil, err
	}
	return returnedSDB, nil
}

// Delete deletes the Safe Deposit Box with the given ID
func (s *SDB) Delete(id string) error {
	id = strings.TrimSpace(id)
	// Check to make sure the ID isn't empty
	if id == "" {
		return ErrorSafeDepositBoxNotFound
	}
	resp, err := s.c.DoRequest(http.MethodDelete, sdbBasePath+"/"+id, map[string]string{}, nil)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		// Check if it is a bad request (improperly set params)
		if resp != nil {
			if resp.StatusCode == http.StatusNotFound {
				return ErrorSafeDepositBoxNotFound
			}
			apiErr := utils.ParseAPIError(resp.Body)
			if apiErr == ErrorBodyNotReturned {
				return fmt.Errorf("Error while deleting SDB. Got HTTP status code %d. %v", resp.StatusCode, apiErr)
			}
			return apiErr
		}
		return fmt.Errorf("Error while deleting SDB: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		apiErr := utils.ParseAPIError(resp.Body)
		if apiErr == ErrorBodyNotReturned {
			return fmt.Errorf("Error while deleting SDB. Got HTTP status code %d. %v", resp.StatusCode, apiErr)
		}
		return apiErr
	}
	return nil
}
