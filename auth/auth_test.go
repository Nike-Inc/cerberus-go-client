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

package auth

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/Nike-Inc/cerberus-go-client/v2/api"
	. "github.com/smartystreets/goconvey/convey"
)

var authResponseBody = `{
    "status": "success",
    "data": {
        "client_token": {
            "client_token": "a-cool-token",
            "policies": [
                "web",
                "stage"
            ],
            "metadata": {
                "username": "john.doe@nike.com",
                "is_admin": "false",
                "groups": "Lst-CDT.CloudPlatformEngine.FTE,Lst-digital.platform-tools.internal"
            },
            "lease_duration": 3600,
            "renewable": true
        }
    }
}`

var expectedResponse = &api.UserAuthResponse{
	Status: api.AuthUserSuccess,
	Data: api.UserAuthData{
		ClientToken: api.UserClientToken{
			ClientToken: "a-cool-token",
			Policies: []string{
				"web",
				"stage",
			},
			Metadata: api.UserMetadata{
				Username: "john.doe@nike.com",
				IsAdmin:  "false",
				Groups:   "Lst-CDT.CloudPlatformEngine.FTE,Lst-digital.platform-tools.internal",
			},
			Duration:  3600,
			Renewable: true,
		},
	},
}

func TestingServer(returnCode int, expectedPath, expectedMethod, body string,
	expectedHeaders map[string]string, f func(ts *httptest.Server)) func() {
	return func() {
		Convey("http requests should be correct", func(c C) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				c.So(r.Method, ShouldEqual, expectedMethod)
				c.So(r.URL.Path, ShouldStartWith, expectedPath)
				// Make sure all expected headers are there
				for k := range expectedHeaders {
					c.So(r.Header.Get(k), ShouldNotBeEmpty)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(returnCode)
				w.Write([]byte(body))
			}))
			f(ts)
			Reset(func() {
				ts.Close()
			})
		})

	}
}

func TestRefresh(t *testing.T) {
	var testToken = "a-test-token"
	var expectedHeaders = map[string]string{
		"X-Cerberus-Token": testToken,
	}
	testHeaders := http.Header{}
	testHeaders.Add("X-Cerberus-Token", testToken)
	Convey("A valid refresh request", t, TestingServer(http.StatusOK, "/v2/auth/user/refresh", http.MethodGet, authResponseBody, expectedHeaders, func(ts *httptest.Server) {
		u, _ := url.Parse(ts.URL)
		Convey("Should not error", func() {
			resp, err := Refresh(*u, testHeaders)
			So(err, ShouldBeNil)
			Convey("And should return a valid auth response", func() {
				So(resp, ShouldResemble, expectedResponse)
			})
		})
	}))

	Convey("An invalid refresh request", t, TestingServer(http.StatusUnauthorized, "/v2/auth/user/refresh", http.MethodGet, "", expectedHeaders, func(ts *httptest.Server) {
		u, _ := url.Parse(ts.URL)
		Convey("Should error", func() {
			resp, err := Refresh(*u, testHeaders)
			So(err, ShouldEqual, api.ErrorUnauthorized)
			So(resp, ShouldBeNil)
		})
	}))

	Convey("A refresh request to an non-responsive server", t, func() {
		u, _ := url.Parse("http://127.0.0.1:32876")
		Convey("Should return an error", func() {
			resp, err := Refresh(*u, testHeaders)
			So(err, ShouldNotBeNil)
			So(resp, ShouldBeNil)
		})
	})
}

func TestLogout(t *testing.T) {
	var testToken = "a-test-token"
	var expectedHeaders = map[string]string{
		"X-Cerberus-Token": testToken,
	}
	testHeaders := http.Header{}
	testHeaders.Add("X-Cerberus-Token", testToken)
	Convey("A valid logout request", t, TestingServer(http.StatusNoContent, "/v1/auth", http.MethodDelete, "", expectedHeaders, func(ts *httptest.Server) {
		u, _ := url.Parse(ts.URL)
		Convey("Should not error", func() {
			err := Logout(*u, testHeaders)
			So(err, ShouldBeNil)
		})
	}))

	Convey("An invalid logout request", t, TestingServer(http.StatusUnauthorized, "/v1/auth", http.MethodDelete, "", expectedHeaders, func(ts *httptest.Server) {
		u, _ := url.Parse(ts.URL)
		Convey("Should error", func() {
			err := Logout(*u, testHeaders)
			So(err, ShouldNotBeNil)
		})
	}))

	Convey("A logout request to an non-responsive server", t, func() {
		u, _ := url.Parse("http://127.0.0.1:32876")
		Convey("Should return an error", func() {
			err := Logout(*u, testHeaders)
			So(err, ShouldNotBeNil)
		})
	})
}
