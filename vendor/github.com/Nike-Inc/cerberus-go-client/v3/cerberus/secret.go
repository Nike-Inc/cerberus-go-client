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
	vault "github.com/hashicorp/vault/api"
)

// Note: This is not tested because it is a simple wrapper on top of Vault, which has its own tests

// Secret wraps the vault.Logical client to make sure all paths are prefaced
// with "secret". This does not expose Unwrap because it will not work with
// Cerberus' path routing
type Secret struct {
	v *vault.Logical
}

const pathPrefix = "secret/"

// Delete deletes the given path. Path should not be prefaced with a "/"
func (s *Secret) Delete(path string) (*vault.Secret, error) {
	return s.v.Delete(pathPrefix + path)
}

// List lists secrets at the given path. Path should not be prefaced with a "/"
func (s *Secret) List(path string) (*vault.Secret, error) {
	return s.v.List(pathPrefix + path)
}

// Read returns the secret at the given path. Path should not be prefaced with a "/"
func (s *Secret) Read(path string) (*vault.Secret, error) {
	return s.v.Read(pathPrefix + path)
}

// Write creates a new secret at the given path. Path should not be prefaced with a "/"
func (s *Secret) Write(path string, data map[string]interface{}) (*vault.Secret, error) {
	return s.v.Write(pathPrefix+path, data)
}
