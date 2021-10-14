# Cerberus Go Client
A Golang client for accessing Cerberus. To learn more about Cerberus, please visit
the [Cerberus Website](http://engineering.nike.com/cerberus/).

![Build](https://github.com/Nike-Inc/cerberus-go-client/workflows/Build/badge.svg?branch=master)
[![GoDoc](https://godoc.org/github.com/Nike-Inc/cerberus-go-client/cerberus?status.svg)](https://godoc.org/github.com/Nike-Inc/cerberus-go-client/cerberus)
[![Go Report Card](https://goreportcard.com/badge/github.com/Nike-Inc/cerberus-go-client)](https://goreportcard.com/report/github.com/Nike-Inc/cerberus-go-client)
[![codecov](https://codecov.io/gh/Nike-Inc/cerberus-go-client/branch/master/graph/badge.svg?token=i9o0gedFAJ)](https://codecov.io/gh/Nike-Inc/cerberus-go-client)
## Usage

### Quick Start
The simplest way to get started is to use the user authentication:
```go
import (
	"fmt"
	"github.com/Nike-Inc/cerberus-go-client/v3/cerberus"
	"github.com/Nike-Inc/cerberus-go-client/v3/auth"
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
Cerberus supports three types of authentication, which are explained below. The authentication types
are designed to be used independently of the full Cerberus client if desired. This allows one
to just get a token for use in other applications. There are also methods for returning a
set of headers needed to authenticate to Cerberus. With all of the authentication types, `GetToken`
triggers the actual authentication process for the given type.

#### STS
STS authentication expects a Cerberus URL and an AWS region in order to authenticate.

```go
authMethod, _ := auth.NewSTSAuth("https://cerberus.example.com", "us-west-2")
token, err := authMethod.GetToken(nil)
```

#### Token
Token authentication is meant to be used when there is already an existing Cerberus token you
wish to use. No validation is done on the token, so if it is invalid or expired, method calls
will likely return an `api.ErrorUnauthorized`.

```go
authMethod, _ := auth.NewTokenAuth("https://cerberus.example.com", "token")
token, err := authMethod.GetToken(nil)
```

### Client
Once you have an authentication method, you can pass it to `NewClient` along with an optional file argument
from which to read the MFA token from. `NewClient` will take care of actually authenticating to Cerberus.

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

For full information on every method, see the [Godoc]().

## Development

### Developing for GOPATH mode (For modifying versions pre v3.0.0)
You must include the source code for this project in the src directory of your $GOPATH. Changes to the GOPATH mode client
must be made in the root directories (cerberus-go-client/api, cerberus-go-client/) 

### Developing for GO Module mode (Modifying versions v3.0.0 and newer)
In version v3.0.0 the Cerberus-Go-Client was migrated to Go Modules for dependency management which removes the requirement
that your code live in the $GOPATH/src directory. Major versions starting at v3.0.0 will now occupy sub directories at the 
project root. 

#### Bumping the major version
If you need to bump the major version please create a new subdirectory called vN (where N is the new major
version, i.e. v3 for v3.0.0) and make a copy of the go client in that directory (copying /api/*.go, /auth/*.go, 
/cerberus/*.go, /utils/*.go, go.mod and Makefiles into the new directory). Update your go.mod file to reflect the new 
major version and update all the imports of the files you copied. 

See https://blog.golang.org/v2-go-modules for more details on this process

### Releasing to go client
Once your code is committed, merge to the master branch tag with the correct symantec versioning for the files you modified. 
If you modified files in the /v3/ directory tag using the correct v3.x.x tag. If you update the go files in the root 
directories tag create a tag with the correct v2.x.x tag.  

### Run Integration Tests

First, make sure the following environment variables are set before running the Go Client integration tests:

``` bash
    export TEST_CERBERUS_URL=https://example.cerberus.com
    export TEST_REGION=us-west-2
    export IAM_PRINCIPAL=arn:aws:iam::111111111:role/example-role
    export USER_GROUP=example.user.group
```

Then, make sure AWS credentials have been obtained. One method is by running [gimme-aws-creds](https://github.com/Nike-Inc/gimme-aws-creds):

```bash
    gimme-aws-creds
```

Next, in the project directory run:
```bash
    go test cerberus-go-client/integration
```

### Known limitations
Currently, this will only support one enrolled MFA device (the first one you enable). In the
future we want to make this cleaner for CLI usage.

## Full example
Below is a full, runnable example of how to use the Cerberus client with a simple CLI

```go
package main

import (
	"fmt"
	"os"

	"github.com/Nike-Inc/cerberus-go-client/v3/cerberus"
	"github.com/Nike-Inc/cerberus-go-client/v3/api"
	"github.com/Nike-Inc/cerberus-go-client/v3/auth"
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

