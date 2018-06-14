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
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path"

	"github.com/Nike-Inc/cerberus-go-client/api"
)

// SecureFile is a subclient for secure files
type SecureFile struct {
	c *Client
}

var secureFileBasePath = "/v1/secure-file"
var secureFileListBasePath = "/v1/secure-files"

// List returns a list of secure files
func (r *SecureFile) List(rootpath string) (*api.SecureFilesResponse, error) {
	resp, err := r.c.DoRequest(http.MethodGet,
		// path.Join will remove last '/' but cerberus expect a / suffix => Let's add it
		path.Join(secureFileListBasePath, rootpath)+"/",
		map[string]string{
			"list": "true",
		},
		nil)
	if err != nil {
		return nil, fmt.Errorf("error while trying to get secure files: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error while trying to list secure files. Got HTTP status code %d",
			resp.StatusCode)
	}
	sfr := &api.SecureFilesResponse{}
	err = parseResponse(resp.Body, sfr)
	if err != nil {
		return nil, err
	}
	return sfr, nil
}

// Get downloads a secure file under localfile. File will be saved in output
func (r *SecureFile) Get(secureFilePath string, output io.Writer) error {
	resp, err := r.c.DoRequest(http.MethodGet,
		path.Join(secureFileBasePath, secureFilePath),
		map[string]string{},
		nil)
	if err != nil {
		return fmt.Errorf("error while downloading secure file: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error while trying to download secure file %s. Got HTTP status code %d",
			secureFilePath,
			resp.StatusCode)
	}

	// Copy
	_, err = io.Copy(output, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

// getUploadFileBodyWriter create a reader containing an encoded multipart file. It returns a reader, a content-type and/or possible error
func getUploadFileBodyWriter(filename string, input io.Reader) (io.Reader, string, error) {
	// Create mpart
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	part, err := w.CreateFormFile("file-content", filename)
	if err != nil {
		return nil, "", err
	}
	// Copy file
	if _, err := io.Copy(part, input); err != nil {
		return nil, "", err
	}

	// save content type of the body
	contentType := w.FormDataContentType()

	// close to flush mpart
	if err := w.Close(); err != nil {
		return nil, "", err
	}

	return &b, contentType, nil
}

// Put uploads a secure file to a given location localfile
func (r *SecureFile) Put(secureFilePath string, filename string, input io.Reader) error {
	// Create multipart body and content type
	body, contentType, err := getUploadFileBodyWriter(filename, input)
	if err != nil {
		return fmt.Errorf("error creating upload body: %v", err)
	}

	// Send request
	resp, err := r.c.DoRequestWithBody(http.MethodPost,
		path.Join(secureFileBasePath, secureFilePath),
		map[string]string{},
		contentType,
		body)
	if err != nil {
		return fmt.Errorf("error while downloading secure file: %v", err)
	}
	defer resp.Body.Close()

	// expected sucess reply is "no content"
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("error while trying to download secure file %s. Got HTTP status code %d",
			secureFilePath,
			resp.StatusCode)
	}

	return nil
}
