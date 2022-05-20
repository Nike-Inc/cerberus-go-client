/*
Copyright 2019 Nike Inc.

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

package integration

import (
	"github.com/Nike-Inc/cerberus-go-client/v3/auth"
	"github.com/Nike-Inc/cerberus-go-client/v3/cerberus"
	"github.com/google/go-cmp/cmp"
	"github.com/satori/go.uuid"
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"testing"
)

func TestClient(t *testing.T) {
	Convey("The Cerberus Go Client", t, func() {
		var region = os.Getenv("TEST_REGION")
		if region == "" {
			t.Error("TEST_REGION must be set as an environment variable")
		}

		var cerberusUrl = os.Getenv("TEST_CERBERUS_URL")
		if cerberusUrl == "" {
			t.Error("TEST_CERBERUS_URL must be set as an environment variable")
		}

		var iamPrincipal = os.Getenv("IAM_PRINCIPAL")
		if iamPrincipal == "" {
			t.Error("IAM_PRINCIPAL must be set as an environment variable")
		}

		var sdbPath = os.Getenv("INTEGRATION_TEST_SDB_PATH")
		if iamPrincipal == "" {
			t.Error("INTEGRATION_TEST_SDB_PATH must be set as an environment variable")
		}

		Convey("Should authenticate with STS Auth", func() {
			authMethod, authErr := auth.NewSTSAuth(cerberusUrl, region)
			So(authErr, ShouldBeNil)
			So(authMethod, ShouldNotBeNil)
			token, tokenErr := authMethod.GetToken(nil)
			So(tokenErr, ShouldBeNil)
			So(token, ShouldNotBeNil)

			Convey("And should create a new client", func() {
				client, clientErr := cerberus.NewClient(authMethod, nil)
				So(clientErr, ShouldBeNil)
				So(client, ShouldNotBeNil)
				tok, getTokenErr := client.Authentication.GetToken(nil)
				So(tok, ShouldEqual, token)
				So(getTokenErr, ShouldBeNil)

				Convey("And should write a secret", func() {
					uuid, _ := uuid.NewV4()
					path := sdbPath + "/secret-payload-" + uuid.String()
					testSecretPayload := map[string]interface{}{"bar": "bop"}
					_, writeErr := client.Secret().Write(path, testSecretPayload)
					So(writeErr, ShouldBeNil)

					Convey("And should read a secret", func() {
						readSecret, readErr := client.Secret().Read(path)
						So(readErr, ShouldBeNil)
						So(cmp.Equal(readSecret.Data, testSecretPayload), ShouldBeTrue)

						Convey("And should delete a secret", func() {
							_, delErr := client.Secret().Delete(path)
							So(delErr, ShouldBeNil)
						})
					})
				})
			})
		})
	})
}
