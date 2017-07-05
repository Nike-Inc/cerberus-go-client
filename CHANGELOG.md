# Changelog / Release Notes

All notable changes to `Cerberus Go Client` will be documented in this file. 

### IAM auth flow change (v0.3.0) - July 2017
This release contains a change in the IAM authentication flow. Instead of refreshing
the token using the refresh endpoint, this changes it to reauthenticate when
`Refresh` is called. This change was made so that automated tools can more easily
use the client.

### Struct name fix (v0.2.2) - June 2017
Contains a fix for a API name change in v2

### IAM auth fix (v0.2.1) - June 2017
This is version contains a fix for the IAM authentication method. The IAM auth
was not handling decryption of the access token using KMS. This release adds
the missing decode step

### Package refactor (v0.2.0) - June 2017

- Puts top level code into a new `cerberus` subdirectory

### Initial Release - May 2017

- Initial code drop for the Cerberus Go client.
  - [Taylor Thomas](https://github.com/thomastaylor312)
  - [Roger Ignazio](https://github.com/rji)
