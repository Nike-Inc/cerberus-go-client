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
	"testing"

	"github.com/Nike-Inc/cerberus-go-client/v2/api"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNewTokenAuth(t *testing.T) {
	Convey("A valid URL and token", t, func() {
		tok, err := NewTokenAuth("https://test.example.com", "kylo-ren")
		Convey("Should return a valid TokenAuth", func() {
			So(err, ShouldBeNil)
			So(tok, ShouldNotBeNil)
			Convey("And should have a valid URL and token", func() {
				So(tok.baseURL.String(), ShouldEqual, "https://test.example.com")
				So(tok.token, ShouldEqual, "kylo-ren")
			})
		})
	})

	Convey("An empty URL", t, func() {
		tok, err := NewTokenAuth("", "tie-fighter")
		Convey("Should error", func() {
			So(err, ShouldNotBeNil)
			So(tok, ShouldBeNil)
		})
	})

	Convey("An empty token", t, func() {
		tok, err := NewTokenAuth("https://test.example.com", "")
		Convey("Should error", func() {
			So(err, ShouldNotBeNil)
			So(tok, ShouldBeNil)
		})
	})

	Convey("An invalid URL", t, func() {
		tok, err := NewTokenAuth("https://test.example.com/a/path", "traitor")
		Convey("Should error", func() {
			So(err, ShouldNotBeNil)
			So(tok, ShouldBeNil)
		})
	})
}

func TestGetToken(t *testing.T) {
	Convey("A valid TokenAuth", t, func() {
		tok, err := NewTokenAuth("https://test.example.com", "rey")
		So(err, ShouldBeNil)
		So(tok, ShouldNotBeNil)
		Convey("Should return a valid token", func() {
			token, err := tok.GetToken(nil)
			So(err, ShouldBeNil)
			So(token, ShouldEqual, "rey")
		})
	})

	Convey("A logged out TokenAuth", t, func() {
		tok, err := NewTokenAuth("https://test.example.com", "rey")
		So(err, ShouldBeNil)
		So(tok, ShouldNotBeNil)
		tok.token = ""
		Convey("Should return an error when getting token", func() {
			token, err := tok.GetToken(nil)
			So(err, ShouldEqual, api.ErrorUnauthenticated)
			So(token, ShouldBeEmpty)
		})
	})
}

func TestIsAuthenticatedToken(t *testing.T) {
	Convey("A valid TokenAuth", t, func() {
		tok, err := NewTokenAuth("https://test.example.com", "rey")
		So(err, ShouldBeNil)
		So(tok, ShouldNotBeNil)
		Convey("Should return true", func() {
			So(tok.IsAuthenticated(), ShouldBeTrue)
		})
	})

	Convey("A logged out TokenAuth", t, func() {
		tok, err := NewTokenAuth("https://test.example.com", "rey")
		So(err, ShouldBeNil)
		So(tok, ShouldNotBeNil)
		tok.token = ""
		Convey("Should return false", func() {
			So(tok.IsAuthenticated(), ShouldBeFalse)
		})
	})
}

func TestRefreshToken(t *testing.T) {
	var testToken = "finn"
	var expectedHeaders = map[string]string{
		"X-Cerberus-Token":  testToken,
		"X-Cerberus-Client": api.ClientHeader,
	}
	testHeaders := http.Header{}
	testHeaders.Add("X-Cerberus-Token", testToken)
	testHeaders.Add("X-Cerberus-Client", api.ClientHeader)
	Convey("A valid TokenAuth", t, TestingServer(http.StatusOK, "/v2/auth/user/refresh", http.MethodGet, authResponseBody, expectedHeaders, func(ts *httptest.Server) {
		tok, err := NewTokenAuth(ts.URL, testToken)
		So(err, ShouldBeNil)
		Convey("Should not error on refresh", func() {
			err := tok.Refresh()
			So(err, ShouldBeNil)
			Convey("And should have a valid new token", func() {
				// See the authResponseBody definition for the location of the new token
				So(tok.token, ShouldEqual, "a-cool-token")
			})
		})
	}))

	Convey("A valid TokenAuth", t, TestingServer(http.StatusInternalServerError, "/v2/auth/user/refresh", http.MethodGet, "", expectedHeaders, func(ts *httptest.Server) {
		tok, err := NewTokenAuth(ts.URL, testToken)
		So(err, ShouldBeNil)
		Convey("Should error with invalid response from server", func() {
			err := tok.Refresh()
			So(err, ShouldNotBeNil)
		})
	}))

	Convey("A logged out TokenAuth", t, func() {
		tok, err := NewTokenAuth("https://test.example.com", "luke")
		So(err, ShouldBeNil)
		So(tok, ShouldNotBeNil)
		tok.token = ""
		Convey("Should error", func() {
			So(tok.Refresh(), ShouldNotBeNil)
		})
	})
}

func TestLogoutToken(t *testing.T) {
	var testToken = "bb-8"
	var expectedHeaders = map[string]string{
		"X-Cerberus-Token":  testToken,
		"X-Cerberus-Client": api.ClientHeader,
	}
	testHeaders := http.Header{}
	testHeaders.Add("X-Cerberus-Token", testToken)
	testHeaders.Add("X-Cerberus-Client", api.ClientHeader)
	Convey("A valid TokenAuth", t, TestingServer(http.StatusNoContent, "/v1/auth", http.MethodDelete, "", expectedHeaders, func(ts *httptest.Server) {
		tok, err := NewTokenAuth(ts.URL, testToken)
		So(err, ShouldBeNil)
		Convey("Should not error on logout", func() {
			err := tok.Logout()
			So(err, ShouldBeNil)
			Convey("And should have an empty token", func() {
				So(tok.token, ShouldBeEmpty)
			})
		})
	}))

	Convey("A valid TokenAuth", t, TestingServer(http.StatusInternalServerError, "/v1/auth", http.MethodDelete, "", expectedHeaders, func(ts *httptest.Server) {
		tok, err := NewTokenAuth(ts.URL, testToken)
		So(err, ShouldBeNil)
		Convey("Should error with invalid response from server", func() {
			err := tok.Logout()
			So(err, ShouldNotBeNil)
		})
	}))

	Convey("A logged out TokenAuth", t, func() {
		tok, err := NewTokenAuth("https://test.example.com", "r2-d2")
		So(err, ShouldBeNil)
		So(tok, ShouldNotBeNil)
		tok.token = ""
		Convey("Should error on logout", func() {
			So(tok.Logout(), ShouldNotBeNil)
		})
	})
}

func TestGetHeadersToken(t *testing.T) {
	Convey("A valid TokenAuth", t, func() {
		tok, err := NewTokenAuth("https://test.example.com", "x-wing")
		So(err, ShouldBeNil)
		So(tok, ShouldNotBeNil)
		Convey("Should return headers", func() {
			headers, err := tok.GetHeaders()
			So(err, ShouldBeNil)
			So(headers, ShouldNotBeNil)
			So(headers.Get("X-Cerberus-Token"), ShouldContainSubstring, "x-wing")
		})
	})

	Convey("A logged out TokenAuth", t, func() {
		tok, err := NewTokenAuth("https://test.example.com", "falcon")
		So(err, ShouldBeNil)
		So(tok, ShouldNotBeNil)
		tok.token = ""
		Convey("Should return an error when getting headers", func() {
			headers, err := tok.GetHeaders()
			So(err, ShouldEqual, api.ErrorUnauthenticated)
			So(headers, ShouldBeNil)
		})
	})
}

func TestGetURLToken(t *testing.T) {
	Convey("A valid TokenAuth", t, func() {
		tok, err := NewTokenAuth("https://test.example.com", "chewbacca")
		So(err, ShouldBeNil)
		So(tok, ShouldNotBeNil)
		Convey("Should return a URL", func() {
			So(tok.GetURL(), ShouldNotBeNil)
			So(tok.GetURL().String(), ShouldEqual, "https://test.example.com")
		})
	})
}

func TestGetExpiryToken(t *testing.T) {
	Convey("A valid TokenAuth", t, func() {
		tok, err := NewTokenAuth("https://test.example.com", "token")
		So(err, ShouldBeNil)
		So(tok, ShouldNotBeNil)
		Convey("Should return zero value expiry and non-nil error", func() {
			exp, err := tok.GetExpiry()
			So(exp, ShouldBeZeroValue)
			So(err, ShouldNotBeNil)
		})
	})
	Convey("A logged out TokenAuth", t, func() {
		tok, err := NewTokenAuth("https://test.example.com", "token")
		So(err, ShouldBeNil)
		So(tok, ShouldNotBeNil)
		tok.token = ""
		Convey("Should return zero value expiry and non-nil error", func() {
			exp, err := tok.GetExpiry()
			So(exp, ShouldBeZeroValue)
			So(err, ShouldNotBeNil)
		})
	})
}
