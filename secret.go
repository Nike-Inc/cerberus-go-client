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
