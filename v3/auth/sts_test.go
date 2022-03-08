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

package auth

import (
	"github.com/Nike-Inc/cerberus-go-client/v3/api"
	"github.com/aws/aws-sdk-go/aws/credentials"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

var responseBody = `{
    "client_token": "token",
    "policies": [ "foo-bar-read", "lookup-self" ],
    "metadata": {
        "aws_region": "us-west-2",
        "iam_principal_arn": "arn:aws:iam::111111111:role/fake-role",
        "username": "arn:aws:iam::111111111:role/fake-role",
        "is_admin": "false",
        "groups": "registered-iam-principals"
    },
    "lease_duration": 3600,
    "renewable": true
}`

func TestNewSTSAuth(t *testing.T) {
	Convey("A valid URL, arn, and region", t, func() {
		a, err := NewSTSAuth("https://test.example.com", "us-east-1")
		Convey("Should return a valid STSAuth", func() {
			So(err, ShouldBeNil)
			So(a, ShouldNotBeNil)
		})
	})
	Convey("An empty URL", t, func() {
		a, err := NewSTSAuth("", "us-east-1")
		Convey("Should error", func() {
			So(err, ShouldNotBeNil)
			So(a, ShouldBeNil)
		})
	})
	Convey("An empty region", t, func() {
		a, err := NewSTSAuth("https://test.example.com", "")
		Convey("Should error", func() {
			So(err, ShouldNotBeNil)
			So(a, ShouldBeNil)
		})
	})
	Convey("An invalid URL", t, func() {
		a, err := NewSTSAuth("https://test.example.com/a/path", "us-east-1")
		Convey("Should error", func() {
			So(err, ShouldNotBeNil)
			So(a, ShouldBeNil)
		})
	})
}

func TestGetTokenSTS(t *testing.T) {
	Convey("A valid STSAuth", t, TestingServer(http.StatusOK, "/v2/auth/sts-identity",
		http.MethodPost, responseBody, map[string]string{"X-Amz-Date": "date",
			"Authorization": "authorization"}, func(ts *httptest.Server) {
			a, err := NewSTSAuth(ts.URL, "us-west-2")
			a.headers.Set("X-Cerberus-Client", api.ClientHeader)
			So(err, ShouldBeNil)
			So(a, ShouldNotBeNil)

			os.Setenv("AWS_ACCESS_KEY_ID", "access")
			os.Setenv("AWS_SECRET_ACCESS_KEY", "secret")
			Convey("Should not error with getting a token", func() {
				tok, err := a.GetToken(nil)
				So(err, ShouldBeNil)
				Convey("And should have a valid token", func() {
					So(tok, ShouldEqual, "token")
				})
				Convey("And should have a valid expiry time", func() {
					So(a.expiry, ShouldHappenOnOrBefore, time.Now().Add(1*time.Hour))
				})
			})
		}))
	Convey("A valid STSAuth", t, TestingServer(http.StatusOK, "/v2/auth/sts-identity",
		http.MethodPost, "{", map[string]string{"X-Amz-Date": "date",
			"Authorization": "authorization"}, func(ts *httptest.Server) {
			a, err := NewSTSAuth(ts.URL, "us-west-2")
			So(err, ShouldBeNil)
			So(a, ShouldNotBeNil)
			Convey("Should error with an invalid response from Cerberus", func() {
				tok, err := a.GetToken(nil)
				So(tok, ShouldBeEmpty)
				So(err, ShouldNotBeNil)
			})
		}))
	Convey("A valid STSAuth", t, func() {
		a, err := NewSTSAuth("https://test.example.com", "us-west-2")
		So(err, ShouldBeNil)
		So(a, ShouldNotBeNil)
		a.expiry = time.Now().Add(100 * time.Second)
		a.token = "test-token"
		Convey("Should return a token if one is set", func() {
			tok, err := a.GetToken(nil)
			So(err, ShouldBeNil)
			So(tok, ShouldEqual, "test-token")
		})
	})
	Convey("A valid STSAuth", t, TestingServer(http.StatusUnauthorized, "/v2/auth/sts-identity",
		http.MethodPost, "", map[string]string{"X-Amz-Date": "date",
			"Authorization": "authorization"}, func(ts *httptest.Server) {
			a, err := NewSTSAuth(ts.URL, "us-west-2")
			So(err, ShouldBeNil)
			So(a, ShouldNotBeNil)

			os.Setenv("AWS_ACCESS_KEY_ID", "access")
			os.Setenv("AWS_SECRET_ACCESS_KEY", "secret")
			Convey("Should error with invalid login", func() {
				tok, err := a.GetToken(nil)
				So(err, ShouldNotBeNil)
				So(tok, ShouldBeEmpty)
			})
		}))
	Convey("A valid STSAuth", t, TestingServer(http.StatusInternalServerError, "/v2/auth/sts-identity",
		http.MethodPost, "", map[string]string{}, func(ts *httptest.Server) {
			a, err := NewSTSAuth(ts.URL, "us-west-2")
			So(err, ShouldBeNil)
			So(a, ShouldNotBeNil)
			Convey("Should error with bad API response", func() {
				tok, err := a.GetToken(nil)
				So(err, ShouldNotBeNil)
				So(tok, ShouldBeEmpty)
			})
		}))
	Convey("An STSAuth with an invalid region", t, TestingServer(http.StatusOK, "/v2/auth/sts-identity",
		http.MethodPost, "{", map[string]string{"X-Amz-Date": "date", "X-Amz-Security-Token": "token",
			"Authorization": "authorization"}, func(ts *httptest.Server) {
			a, err := NewSTSAuth(ts.URL, "test-region")
			So(err, ShouldBeNil)
			So(a, ShouldNotBeNil)
			Convey("Should error", func() {
				tok, err := a.GetToken(nil)
				So(tok, ShouldBeEmpty)
				So(err, ShouldNotBeNil)
			})
		}))
}

func TestGetExpiry(t *testing.T) {
	Convey("A valid STSAuth", t, func() {
		a, err := NewSTSAuth("https://test.example.com", "us-west-2")
		So(err, ShouldBeNil)
		So(a, ShouldNotBeNil)
		a.expiry = time.Now()
		a.token = "token"
		Convey("Should return an expiry time", func() {
			exp, err := a.GetExpiry()
			So(exp, ShouldNotBeNil)
			So(err, ShouldBeNil)
		})
	})
	Convey("An unauthenticated STSAuth", t, func() {
		a, err := NewSTSAuth("https://test.example.com", "us-west-2")
		So(err, ShouldBeNil)
		So(a, ShouldNotBeNil)
		Convey("Should return an error", func() {
			exp, err := a.GetExpiry()
			So(exp, ShouldBeZeroValue)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestIsAuthenticated(t *testing.T) {
	Convey("A valid STSAuth", t, func() {
		a, err := NewSTSAuth("https://test.example.com", "us-west-2")
		So(err, ShouldBeNil)
		So(a, ShouldNotBeNil)
		a.expiry = time.Now().Add(100 * time.Second)
		a.token = "token"
		Convey("Should return true", func() {
			So(a.IsAuthenticated(), ShouldBeTrue)
		})
	})
	Convey("An unauthenticated STSAuth", t, func() {
		a, err := NewSTSAuth("https://test.example.com", "us-west-2")
		So(err, ShouldBeNil)
		So(a, ShouldNotBeNil)
		Convey("Should return false", func() {
			So(a.IsAuthenticated(), ShouldBeFalse)
		})
	})
}

func TestRefreshSTS(t *testing.T) {
	Convey("An unauthenticated STSAuth", t, func() {
		a, err := NewSTSAuth("https://test.example.com", "us-west-2")
		So(err, ShouldBeNil)
		So(a, ShouldNotBeNil)
		Convey("Should error", func() {
			So(a.Refresh(), ShouldEqual, api.ErrorUnauthenticated)
		})
	})
}

func TestLogoutSTS(t *testing.T) {
	var testToken = "token"
	var expectedHeaders = map[string]string{
		"X-Cerberus-Token":  testToken,
		"X-Cerberus-Client": api.ClientHeader,
	}
	Convey("A valid STSAuth", t, TestingServer(http.StatusNoContent, "/v1/auth", http.MethodDelete,
		"", expectedHeaders, func(ts *httptest.Server) {
			testHeaders := http.Header{}
			testHeaders.Add("X-Cerberus-Token", testToken)
			testHeaders.Add("X-Cerberus-Client", api.ClientHeader)
			a, err := NewSTSAuth(ts.URL, "us-west-2")
			So(err, ShouldBeNil)
			So(a, ShouldNotBeNil)
			a.expiry = time.Now().Add(100 * time.Second)
			a.token = testToken
			a.headers = testHeaders
			Convey("Should not error on logout", func() {
				err := a.Logout()
				So(err, ShouldBeNil)
				Convey("And should have an empty token", func() {
					So(a.token, ShouldBeEmpty)
				})
			})
		}))
	Convey("A valid STSAuth", t, TestingServer(http.StatusInternalServerError, "/v1/auth",
		http.MethodDelete, "", expectedHeaders, func(ts *httptest.Server) {
			testHeaders := http.Header{}
			testHeaders.Add("X-Cerberus-Token", testToken)
			testHeaders.Add("X-Cerberus-Client", api.ClientHeader)
			a, err := NewSTSAuth(ts.URL, "us-west-2")
			So(err, ShouldBeNil)
			So(a, ShouldNotBeNil)
			a.expiry = time.Now().Add(100 * time.Second)
			a.token = testToken
			a.headers = testHeaders
			Convey("Should error with invalid response from server", func() {
				err := a.Logout()
				So(err, ShouldNotBeNil)
			})
		}))
	Convey("An unauthenticated STSAuth", t, func() {
		a, err := NewSTSAuth("https://test.example.com", "us-west-2")
		So(err, ShouldBeNil)
		So(a, ShouldNotBeNil)
		Convey("Should error on logout", func() {
			So(a.Logout(), ShouldEqual, api.ErrorUnauthenticated)
		})
	})
}

func TestGetHeaders(t *testing.T) {
	var testToken = "token"
	testHeaders := http.Header{}
	testHeaders.Add("X-Cerberus-Token", testToken)
	Convey("A valid STSAuth", t, func() {
		a, err := NewSTSAuth("https://test.example.com", "us-west-2")
		So(err, ShouldBeNil)
		So(a, ShouldNotBeNil)
		a.expiry = time.Now().Add(100 * time.Second)
		a.token = testToken
		a.headers = testHeaders
		Convey("Should return headers", func() {
			headers, err := a.GetHeaders()
			So(err, ShouldBeNil)
			So(headers, ShouldNotBeNil)
			So(headers.Get("X-Cerberus-Token"), ShouldContainSubstring, testToken)
		})
	})
	Convey("An unauthenticated STSAuth", t, func() {
		a, err := NewSTSAuth("https://test.example.com", "us-west-2")
		So(err, ShouldBeNil)
		So(a, ShouldNotBeNil)
		Convey("Should return an error when getting headers", func() {
			headers, err := a.GetHeaders()
			So(err, ShouldEqual, api.ErrorUnauthenticated)
			So(headers, ShouldBeNil)
		})
	})
}

func TestGetURL(t *testing.T) {
	Convey("A valid STSAuth", t, func() {
		a, err := NewSTSAuth("https://test.example.com", "us-east-1")
		So(err, ShouldBeNil)
		So(a, ShouldNotBeNil)
		Convey("Should return a URL", func() {
			So(a.GetURL(), ShouldNotBeNil)
			So(a.GetURL().String(), ShouldEqual, "https://test.example.com")
		})
	})
}

func TestCreds(t *testing.T) {
	Convey("When credentials exist, a call", t, func() {
		value := credentials.Value{AccessKeyID: "access", SecretAccessKey: "secret", SessionToken: "session",
			ProviderName: "provider"}
		c := credentials.NewStaticCredentialsFromCreds(value)
		Convey("Should have credentials", func() {
			creds, err := c.Get()
			So(err, ShouldBeNil)
			So(creds, ShouldNotBeNil)
		})
	})
}

func TestSigner(t *testing.T) {
	Convey("A signer with credentials", t, func() {
		os.Setenv("AWS_ACCESS_KEY_ID", "access")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "secret")
		a, e := signer(creds)
		Convey("Should return a signer", func() {
			So(a, ShouldNotBeNil)
			So(e, ShouldBeNil)
		})
	})
}

func TestRequest(t *testing.T) {
	Convey("A valid request call", t, func() {
		a, err := NewSTSAuth("https://test.example.com", "us-west-2")
		So(err, ShouldBeNil)
		So(a, ShouldNotBeNil)
		r, e := a.request()
		Convey("Should return a request", func() {
			So(e, ShouldBeNil)
			So(r.Method, ShouldEqual, "POST")
			So(r.Host, ShouldEqual, "sts.us-west-2.amazonaws.com")
		})
	})
	Convey("A request call with an invalid region", t, func() {
		a, err := NewSTSAuth("https://test.example.com", "test-region")
		So(err, ShouldBeNil)
		So(a, ShouldNotBeNil)
		r, e := a.request()
		Convey("Should error", func() {
			So(e, ShouldNotBeNil)
			So(r, ShouldBeNil)
		})
	})
	Convey("A valid request call to cn-north-1", t, func() {
		a, err := NewSTSAuth("https://test.example.com", "cn-north-1")
		So(err, ShouldBeNil)
		So(a, ShouldNotBeNil)
		r, e := a.request()
		Convey("Should return a request", func() {
			So(e, ShouldBeNil)
			So(r.Method, ShouldEqual, "POST")
			So(r.Host, ShouldEqual, "sts.cn-north-1.amazonaws.com.cn")
		})
	})
	Convey("A valid request call to cn-northwest-1", t, func() {
		a, err := NewSTSAuth("https://test.example.com", "cn-northwest-1")
		So(err, ShouldBeNil)
		So(a, ShouldNotBeNil)
		r, e := a.request()
		Convey("Should return a request", func() {
			So(e, ShouldBeNil)
			So(r.Method, ShouldEqual, "POST")
			So(r.Host, ShouldEqual, "sts.cn-northwest-1.amazonaws.com.cn")
		})
	})
}

func TestSign(t *testing.T) {
	Convey("A valid signing", t, func() {
		a, err := NewSTSAuth("https://test.example.com", "us-west-2")
		So(err, ShouldBeNil)
		So(a, ShouldNotBeNil)

		os.Setenv("AWS_ACCESS_KEY_ID", "access")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "secret")
		r, err := a.sign()
		Convey("Should sign a request", func() {
			So(err, ShouldBeNil)
			So(r.Get("X-Amz-Security-Token"), ShouldNotBeNil)
			So(r.Get("X-Amz-Date"), ShouldNotBeNil)
			So(r.Get("Authorization"), ShouldNotBeNil)
		})
	})
	Convey("A signing with a bad request", t, func() {
		a, err := NewSTSAuth("https://test.example.com", "test-regopm")
		So(err, ShouldBeNil)
		So(a, ShouldNotBeNil)
		r, err := a.sign()
		Convey("Should error", func() {
			So(err, ShouldNotBeNil)
			So(r, ShouldBeNil)
		})
	})
}
