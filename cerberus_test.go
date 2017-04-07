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

	. "github.com/smartystreets/goconvey/convey"
	"github.nike.com/ngp/cerberus-client-go/api"
)

type MockAuth struct {
	baseURL     *url.URL
	headers     http.Header
	token       string
	getTokenErr bool
	refreshErr  bool
}

func GenerateMockAuth(cerberusURL, token string, tokenErr, refreshErr bool) *MockAuth {
	baseURL, _ := url.Parse(cerberusURL)
	return &MockAuth{
		baseURL: baseURL,
		headers: http.Header{
			"Content-Type":  []string{"application/json"},
			"X-Vault-Token": []string{token},
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
	return "", fmt.Errorf("Arrrrrg...an error matey")
}

func (m *MockAuth) IsAuthenticated() bool {
	return len(m.token) > 0
}

func (m *MockAuth) Refresh() error {
	if !m.refreshErr {
		return nil
	}
	return fmt.Errorf("Arrrrrg...an error matey")
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

func WithServer(returnCode int, expectedPath, expectedMethod, bodyContains string, f func(ts *httptest.Server)) func() {
	return func() {
		Convey("http requests should be correct", func(c C) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				c.So(r.Method, ShouldEqual, expectedMethod)
				c.So(r.URL.Path, ShouldStartWith, expectedPath)
				body, err := ioutil.ReadAll(r.Body)
				c.So(err, ShouldBeNil)
				c.So(string(body), ShouldContainSubstring, bodyContains)
				w.Header().Set("Content-Type", "application/json")
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
	Convey("Valid GET request", t, WithServer(http.StatusOK, "/v1/blah", http.MethodGet, "", func(ts *httptest.Server) {
		cl, _ := NewClient(GenerateMockAuth(ts.URL, "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		Convey("Should return a valid response", func() {
			resp, err := cl.DoRequest(http.MethodGet, "/v1/blah", nil)
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, http.StatusOK)
		})
	}))

	Convey("Valid POST request", t, WithServer(http.StatusOK, "/v1/books/armaments", http.MethodPost, "holy hand grenade of antioch", func(ts *httptest.Server) {
		cl, _ := NewClient(GenerateMockAuth(ts.URL, "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		var testData = map[string]string{
			"character": "Brother Maynard",
			"weapon":    "holy hand grenade of antioch",
		}
		Convey("Should return a valid response", func() {
			resp, err := cl.DoRequest(http.MethodPost, "/v1/books/armaments", testData)
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, http.StatusOK)
		})
	}))

	Convey("A request to a non-responsive server", t, func() {
		cl, _ := NewClient(GenerateMockAuth("http://127.0.0.1:32876", "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		Convey("Should return an error", func() {
			resp, err := cl.DoRequest(http.MethodGet, "/v1/blah", nil)
			So(err, ShouldNotBeNil)
			So(resp, ShouldBeNil)
		})
	})
}
