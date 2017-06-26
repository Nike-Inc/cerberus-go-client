# Cerberus Go Client
A Golang client for accessing Cerberus. To learn more about Cerberus, please visit
the [Cerberus Website](http://engineering.nike.com/cerberus/).

[![Build Status](https://travis-ci.org/Nike-Inc/cerberus-go-client.svg?branch=master)](https://travis-ci.org/Nike-Inc/cerberus-go-client)
[![GoDoc](https://godoc.org/github.com/Nike-Inc/cerberus-go-client/cerberus?status.svg)](https://godoc.org/github.com/Nike-Inc/cerberus-go-client/cerberus)
[![Go Report Card](https://goreportcard.com/badge/github.com/Nike-Inc/cerberus-go-client)](https://goreportcard.com/report/github.com/Nike-Inc/cerberus-go-client)

## Usage

### Quick Start
The simplest way to get started is to use the user authentication:
```go
import (
	"fmt"

	"github.com/Nike-Inc/cerberus-go-client/cerberus"
	"github.com/Nike-Inc/cerberus-go-client/auth"
)
...
authMethod, _ := auth.NewUserAuth("https://cerberus.example.com", "my-cerberus-user", "my-password")
// This will prompt you for an MFA token if you have MFA enabled
client, err := cerberus.NewClient(authMethod, nil)
if err != nil {
    panic(err)
}
sdbs, _ := client.SDB().List()
fmt.Println(sdbs)
```

### Supported endpoints
All of the most recent endpoints at the time of writing (2 May 2017) are supported while older
versions are not. The full list is below

- `/v2/auth/user`
- `/v2/auth/mfa_check`
- `/v2/auth/user/refresh`
- `/v2/auth/iam-principal`
- `/v1/auth` (used for `DELETE` operations)
- `/v2/safe-deposit-box`
- `/v1/role`
- `/v1/category`
- `/v1/metadata`

### Authentication
Cerberus supports 3 types of authentication, all of which are explained below. The auth types
are designed to be used independently of the full Cerberus client if desired. This allows you
to just get a token for use in other applications. There are also methods for returning a
set of headers needed to authenticate to Cerberus. With all of the authentication types, `GetToken`
triggers the actual authentication process for the given type.

All 3 types support setting the URL for Cerberus using the `CERBERUS_URL` environment variable,
which will always override anything you pass to the `New*Auth` methods.

#### AWS
AWS authentication expects an IAM principal ARN and an AWS region to be able to authenticate.
For more information, see the [API docs](https://github.com/Nike-Inc/cerberus-management-service/blob/master/API.md#app-login-v2-v2authiam-principal)

```go
authMethod, _ := auth.NewAWSAuth("https://cerberus.example.com", "arn:aws:iam::111111111:role/cerberus-api-tester", "us-west-2")
tok, err := authMethod.GetToken(nil)
```

#### Token
Token authentication is meant to be used when there is already an existing Cerberus token you
wish to use. No validation is done on the token, so if it is invalid or expired, method calls
will likely return an `api.ErrorUnauthorized`.

This method also allows you to set a token using the `CERBERUS_TOKEN` environment variable.
Like `CERBERUS_URL`, this will override anything you pass to the `NewTokenAuth` method.

```go
authMethod, _ := auth.NewTokenAuth("https://cerberus.example.com", "my-cool-token")
tok, err := authMethod.GetToken(nil)
```

#### User
User authentication is for using a username and password (with optional MFA) to log in to Cerberus.
There are some [known limitations](#known-limitations) with MFA. The `GetToken` method takes an `*os.File`
argument that expects a file with one line containing the MFA token to use. Otherwise, if `nil` is passed
it will prompt for the MFA token.

```go
authMethod, _ := auth.NewUserAuth("https://cerberus.example.com", "my-cerberus-user", "my-password")
tok, err := authMethod.GetToken(nil)
```

### Client
Once you have an authentication method, you can pass it to `NewClient` along with an optional file argument
for where to read the MFA token from. `NewClient` will take care of actually authenticating to Cerberus

```go
client, err := cerberus.NewClient(authMethod, nil)
```

The client is organized with various "subclients" to access different endpoints. For example, to list all
SDBs and secrets for each SDB:

```go
list, err := client.SDB().List()
for _, v := range list {
    l, _ := client.Secret().List(v.Path)
    fmt.Println(l)
}
```

For full information on every method, see the [Godoc]()

## Development

### Code organization
The code is broken up into 4 parts, including 3 subpackages. The top level package contains all of
the code for the Cerberus client proper. A breakdown of all the subpackages follows:

#### API
The `api` package contains all type definitions for API objects as well as common errors. It also contains
API error handling methods

#### Auth
The `auth` package contains implementations for all authentication types and the definition for the `Auth`
interface that all authentication types must satisfy.

#### Utils
The `utils` package contains common methods used by the top level client and multiple subpackages. This
**is not** meant to be a kitchen sink in which to throw things that don't belong.

### Tests
We use [GoConvey](https://github.com/smartystreets/goconvey) for our testing. There are plenty of tests
in the code that you can use for examples

### Contributing
See the [CONTRIBUTING.md](CONTRIBUTING.md) document for more information on how to begin contributing.

The tl;dr is that we encourage any PRs you'd like to submit. Please remember to keep your commits
small and focused and try to write tests for any new features you add.

## Roadmap
All endpoints have been implemented with unit tests. The other major task remaining is to write
integration tests.

### Known limitations
Currently, this will only support one enrolled MFA device (the first one you enable). In the
future we want to make this cleaner for CLI usage

## Full example
Below is a full, runnable example of how to use the Cerberus client with a simple CLI

```go
package main

import (
	"fmt"
	"os"

	"github.com/Nike-Inc/cerberus-go-client/cerberus"
	"github.com/Nike-Inc/cerberus-go-client/api"
	"github.com/Nike-Inc/cerberus-go-client/auth"
	"golang.org/x/crypto/ssh/terminal"
)

func main() {
	var username string
	fmt.Print("Input username: ")
	fmt.Scan(&username)
	fmt.Print("Input password: ")
	password, _ := terminal.ReadPassword(0)
	fmt.Print("\n")
	authMethod, err := auth.NewUserAuth("https://cerberus.example.com", username, string(password))
	if err != nil {
		fmt.Printf("Error when creating auth method: %v\n", err)
		os.Exit(1)
	}

	client, err := cerberus.NewClient(authMethod, nil)
	if err != nil {
		fmt.Printf("Error when creating client: %v\n", err)
		os.Exit(1)
	}
	tok, _ := client.Authentication.GetToken(nil)
	fmt.Println(tok)

	sdb, err := client.SDB().GetByName("TestBoxForScience")
	if err != nil {
		fmt.Printf("Error when getting sdb: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(sdb)

	sec, err := client.Secret().List(sdb.Path)
	if err != nil {
		fmt.Printf("Error when getting secrets: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(sec)

	newSDB, err := client.SDB().Create(&api.SafeDepositBox{
		Name:        "testtest",
		Description: "A test thing",
		Owner:       "Lst.test.cerberus",
		CategoryID:  sdb.CategoryID,
	})
	if err != nil {
		fmt.Printf("Error when creating SDB: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(newSDB)

	if err := client.SDB().Delete(newSDB.ID); err != nil {
		fmt.Printf("Error when deleting SDB: %v\n", err)
		os.Exit(1)
	}
}
```

## Maintainers
- [Taylor Thomas](https://github.com/thomastaylor312)
- [Roger Ignazio](https://github.com/rji)
