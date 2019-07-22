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

package cerberus

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/Nike-Inc/cerberus-go-client/api"
	. "github.com/smartystreets/goconvey/convey"
)

type MockAuth struct {
	baseURL     *url.URL
	headers     http.Header
	token       string
	getTokenErr bool
	refreshErr  bool
}

const refreshedToken = "a refreshed token"

func GenerateMockAuth(cerberusURL, token string, tokenErr, refreshErr bool) *MockAuth {
	baseURL, _ := url.Parse(cerberusURL)
	return &MockAuth{
		baseURL: baseURL,
		headers: http.Header{
			"Content-Type":     []string{"application/json"},
			"X-Cerberus-Token": []string{token},
		},
		token:       token,
		getTokenErr: tokenErr,
		refreshErr:  refreshErr,
	}
}

func (m *MockAuth) GetToken(f *os.File) (string, error) {
	if !m.getTokenErr {
		return m.token, nil
	}
	return "", fmt.Errorf("MockAuth unable to obtain token")
}

func (m *MockAuth) IsAuthenticated() bool {
	return len(m.token) > 0
}

func (m *MockAuth) Refresh() error {
	if !m.refreshErr {
		m.token = refreshedToken
		return nil
	}
	return fmt.Errorf("MockAuth unable to obtain token")
}

func (m *MockAuth) Logout() error {
	m.token = ""
	return nil
}

func (m *MockAuth) GetHeaders() (http.Header, error) {
	return m.headers, nil
}

func (m *MockAuth) GetURL() *url.URL {
	return m.baseURL
}

func (m *MockAuth) GetExpiry() (time.Time, error) {
	return time.Now(), nil
}

func TestNewCerberusClient(t *testing.T) {
	Convey("Valid setup arguments", t, func() {
		m := GenerateMockAuth("http://example.com", "a-cool-token", false, false)
		c, err := NewClient(m, nil)
		Convey("Should result in a valid client", func() {
			So(err, ShouldBeNil)
			So(c, ShouldNotBeNil)
		})
	})

	Convey("Bad login to get token", t, func() {
		m := GenerateMockAuth("http://example.com", "a-cool-token", true, false)
		c, err := NewClient(m, nil)
		Convey("Should error", func() {
			So(err, ShouldNotBeNil)
			So(c, ShouldBeNil)
		})
	})
}

func TestSubclients(t *testing.T) {
	Convey("A valid client", t, func() {
		m := GenerateMockAuth("http://example.com", "a-cool-token", false, false)
		c, _ := NewClient(m, nil)
		So(c, ShouldNotBeNil)
		Convey("Should return a valid SDB client", func() {
			So(c.SDB(), ShouldNotBeNil)
		})
		Convey("Should return a valid Secret client", func() {
			So(c.Secret(), ShouldNotBeNil)
		})
		Convey("Should return a valid Role client", func() {
			So(c.Role(), ShouldNotBeNil)
		})
		Convey("Should return a valid Category client", func() {
			So(c.Category(), ShouldNotBeNil)
		})
		Convey("Should return a valid Metadata client", func() {
			So(c.Metadata(), ShouldNotBeNil)
		})
	})
}

func TestParseResponse(t *testing.T) {
	Convey("Valid JSON object", t, func() {
		buf := bytes.NewBuffer([]byte(`{
			"id": "123",
			"name": "IAMObject"
		}`))
		expected := &api.MFADevice{
			ID:   "123",
			Name: "IAMObject",
		}
		obj := &api.MFADevice{}
		err := parseResponse(buf, obj)
		Convey("Should parse correctly", func() {
			So(err, ShouldBeNil)
			So(obj, ShouldResemble, expected)
		})
	})
	Convey("Invalid JSON object", t, func() {
		buf := bytes.NewBuffer([]byte(`{
			"id": 1,
			"name": "IAMObject"
		}`))
		obj := &api.MFADevice{}
		err := parseResponse(buf, obj)
		Convey("Should error", func() {
			So(err, ShouldNotBeNil)
		})
	})
}

func WithServer(returnCode int, shouldRefresh bool, expectedPath, expectedMethod, bodyContains string, expectedParams map[string]string, f func(ts *httptest.Server)) func() {
	return func() {
		Convey("http requests should be correct", func(c C) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				c.So(r.Method, ShouldEqual, expectedMethod)
				c.So(r.URL.Path, ShouldStartWith, expectedPath)
				// Make sure all expected params are there
				for k, v := range expectedParams {
					c.So(r.FormValue(k), ShouldEqual, v)
				}
				body, err := ioutil.ReadAll(r.Body)
				c.So(err, ShouldBeNil)
				c.So(string(body), ShouldContainSubstring, bodyContains)
				w.Header().Set("Content-Type", "application/json")
				if shouldRefresh {
					w.Header().Set("X-Refresh-Token", "true")
				}
				w.WriteHeader(returnCode)
				w.Write([]byte(`{"message": "a message"}`))
			}))
			f(ts)
			Reset(func() {
				ts.Close()
			})
		})
	}
}

func TestDoRequest(t *testing.T) {
	var testParams = map[string]string{
		"theNumberThouShaltCountTo": "3",
		"rightOut":                  "5",
	}
	Convey("Valid GET request", t, WithServer(http.StatusOK, false, "/v1/blah", http.MethodGet, "", map[string]string{}, func(ts *httptest.Server) {
		cl, _ := NewClient(GenerateMockAuth(ts.URL, "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		Convey("Should return a valid response", func() {
			resp, err := cl.DoRequest(http.MethodGet, "/v1/blah", map[string]string{}, nil)
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, http.StatusOK)
		})
	}))

	Convey("Valid request with params", t, WithServer(http.StatusOK, false, "/v1/blah", http.MethodGet, "", testParams, func(ts *httptest.Server) {
		cl, _ := NewClient(GenerateMockAuth(ts.URL, "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		Convey("Should return a valid response", func() {
			resp, err := cl.DoRequest(http.MethodGet, "/v1/blah", testParams, nil)
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, http.StatusOK)
		})
	}))

	Convey("Valid POST request", t, WithServer(http.StatusOK, true, "/v1/books/armaments", http.MethodPost, "holy hand grenade of antioch", map[string]string{}, func(ts *httptest.Server) {
		cl, _ := NewClient(GenerateMockAuth(ts.URL, "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		var testData = map[string]string{
			"character": "Brother Maynard",
			"weapon":    "holy hand grenade of antioch",
		}
		Convey("Should return a valid response", func() {
			resp, err := cl.DoRequest(http.MethodPost, "/v1/books/armaments", map[string]string{}, testData)
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, http.StatusOK)
		})
	}))

	Convey("Valid POST request", t, WithServer(http.StatusOK, true, "/v1/books/armaments", http.MethodPost, "holy hand grenade of antioch", map[string]string{}, func(ts *httptest.Server) {
		cl, _ := NewClient(GenerateMockAuth(ts.URL, "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		var testData = map[string]string{
			"character": "Brother Maynard",
			"weapon":    "holy hand grenade of antioch",
		}
		Convey("Should return a valid response", func() {
			resp, err := cl.DoRequest(http.MethodPost, "/v1/books/armaments", map[string]string{}, testData)
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, http.StatusOK)
			Convey("Vault token should be set to the new token", func() {
				So(cl.vaultClient.Token(), ShouldEqual, refreshedToken)
			})
		})
	}))

	Convey("Valid POST request with failed refresh", t, WithServer(http.StatusOK, true, "/v1/books/armaments", http.MethodPost, "holy hand grenade of antioch", map[string]string{}, func(ts *httptest.Server) {
		cl, _ := NewClient(GenerateMockAuth(ts.URL, "a-cool-token", false, true), nil)
		So(cl, ShouldNotBeNil)
		var testData = map[string]string{
			"character": "Brother Maynard",
			"weapon":    "holy hand grenade of antioch",
		}
		Convey("Should still return a valid response", func() {
			resp, err := cl.DoRequest(http.MethodPost, "/v1/books/armaments", map[string]string{}, testData)
			So(err, ShouldNotBeNil)
			So(resp.StatusCode, ShouldEqual, http.StatusOK)
		})
	}))

	Convey("A request to a non-responsive server", t, func() {
		cl, _ := NewClient(GenerateMockAuth("http://127.0.0.1:32876", "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		Convey("Should return an error", func() {
			resp, err := cl.DoRequest(http.MethodGet, "/v1/blah", map[string]string{}, nil)
			So(err, ShouldNotBeNil)
			So(resp, ShouldBeNil)
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
		err := handleAPIError(buf)
		Convey("Should parse correctly", func() {
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, expected)
		})
	})
	Convey("Empty body", t, func() {
		buf := bytes.NewBuffer([]byte(""))
		err := handleAPIError(buf)
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
		err := handleAPIError(buf)
		Convey("Should have a normal error response", func() {
			So(err, ShouldNotBeNil)
			So(err, ShouldNotHaveSameTypeAs, api.ErrorResponse{})
			So(err, ShouldNotEqual, ErrorBodyNotReturned)
		})
	})
}
