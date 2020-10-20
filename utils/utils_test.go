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

package utils

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Nike-Inc/cerberus-go-client/v3/api"
	. "github.com/smartystreets/goconvey/convey"
)

func TestValidateURL(t *testing.T) {
	Convey("A valid URL", t, func() {
		parsedURL, err := ValidateURL("https://a.cerberus.com:3030")
		Convey("Should not error", func() {
			So(err, ShouldBeNil)
			So(parsedURL, ShouldNotBeNil)
		})
	})

	Convey("An invalid URL", t, func() {
		parsedURL, err := ValidateURL("https://a.cerberus.%com:3030")
		Convey("Should error", func() {
			So(err, ShouldNotBeNil)
			So(parsedURL, ShouldBeNil)
		})
	})

	Convey("A URL with a path", t, func() {
		parsedURL, err := ValidateURL("https://a.cerberus.com/foo/bar/baz")
		Convey("Should error", func() {
			So(err, ShouldNotBeNil)
			So(parsedURL, ShouldBeNil)
		})
	})
	Convey("A URL with a path", t, func() {
		parsedURL, err := ValidateURL("https://a.cerberus.com?i=like&query=params")
		Convey("Should error", func() {
			So(err, ShouldNotBeNil)
			So(parsedURL, ShouldBeNil)
		})
	})
}

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

func TestCheckAndParse(t *testing.T) {
	Convey("A valid response", t, func() {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(authResponseBody))
		}))
		defer ts.Close()
		Convey("Should not error", func() {
			resp, err := http.Get(ts.URL)
			So(err, ShouldBeNil)
			authResp, err := CheckAndParse(resp)
			So(err, ShouldBeNil)
			So(authResp, ShouldNotBeNil)
			Convey("An should have a properly parsed response", func() {
				So(authResp, ShouldResemble, expectedResponse)
			})
		})
	})

	Convey("An invalid body", t, func() {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("{bad json"))
		}))
		defer ts.Close()
		Convey("Should error", func() {
			resp, err := http.Get(ts.URL)
			So(err, ShouldBeNil)
			authResp, err := CheckAndParse(resp)
			So(err, ShouldNotBeNil)
			So(authResp, ShouldBeNil)
		})
	})

	Convey("A forbidden response", t, func() {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(""))
		}))
		defer ts.Close()
		Convey("Should error", func() {
			resp, err := http.Get(ts.URL)
			So(err, ShouldBeNil)
			authResp, err := CheckAndParse(resp)
			So(err, ShouldEqual, api.ErrorUnauthorized)
			So(authResp, ShouldBeNil)
		})
	})

	Convey("A unauthorized response", t, func() {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(""))
		}))
		defer ts.Close()
		Convey("Should error", func() {
			resp, err := http.Get(ts.URL)
			So(err, ShouldBeNil)
			authResp, err := CheckAndParse(resp)
			So(err, ShouldEqual, api.ErrorUnauthorized)
			So(authResp, ShouldBeNil)
		})
	})

	Convey("A server error", t, func() {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(""))
		}))
		defer ts.Close()
		Convey("Should return an error", func() {
			resp, err := http.Get(ts.URL)
			So(err, ShouldBeNil)
			authResp, err := CheckAndParse(resp)
			So(err, ShouldNotBeNil)
			So(authResp, ShouldBeNil)
		})
	})
}

func TestHandleAPIError(t *testing.T) {
	Convey("Valid error body", t, func() {
		buf := bytes.NewBuffer([]byte(`{
	"error_id": "a041aa4d-1d5a-4eed-8e8a-6dc18bdf96db",
	"errors": [{
		"code": 99208,
		"message": "The name may not be blank.",
		"metadata": {
			"field": "name"
		}
	}]
}`))
		expected := api.ErrorResponse{
			ErrorID: "a041aa4d-1d5a-4eed-8e8a-6dc18bdf96db",
			Errors: []api.ErrorDetail{
				api.ErrorDetail{
					Code:    99208,
					Message: "The name may not be blank.",
					Metadata: map[string]interface{}{
						"field": "name",
					},
				},
			},
		}
		err := ParseAPIError(buf)
		Convey("Should parse correctly", func() {
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, expected)
		})
	})
	Convey("Empty body", t, func() {
		buf := bytes.NewBuffer([]byte(""))
		err := ParseAPIError(buf)
		Convey("Should have a normal error response", func() {
			So(err, ShouldNotBeNil)
			So(err, ShouldEqual, ErrorBodyNotReturned)
		})
	})
	Convey("Invalid JSON object", t, func() {
		buf := bytes.NewBuffer([]byte(`{
			"id": 1,
			"name": "weirdobj"
		`))
		err := ParseAPIError(buf)
		Convey("Should have a normal error response", func() {
			So(err, ShouldNotBeNil)
			So(err, ShouldNotHaveSameTypeAs, api.ErrorResponse{})
			So(err, ShouldNotEqual, ErrorBodyNotReturned)
		})
	})
}
