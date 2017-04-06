package auth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"io"

	"os"

	. "github.com/smartystreets/goconvey/convey"
	"github.nike.com/ngp/cerberus-client-go/api"
)

var validLogin = `{
    "status": "%s",
    "data": {
        "client_token": {
            "client_token": "%s",
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

var validLoginMFA = `{
    "status" : "%s",
    "data" : {
      "user_id" : "13427265",
      "username" : "john.doe@nike.com",
      "state_token" : "%s",
      "devices" : [ {
        "id" : "111111",
        "name" : "Google Authenticator"
      }, {
        "id" : "22222",
        "name" : "Google Authenticator"
      }, {
        "id" : "33333",
        "name" : "Google Authenticator"
      } ],
      "client_token" : null
    }
}`

func TestNewUserAuth(t *testing.T) {
	Convey("Bad NewUserAuth setup", t, func() {
		Convey("should error with empty username", func() {
			c, err := NewUserAuth("https://test.example.com", "", "password")
			So(c, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})
		Convey("should error with empty password", func() {
			c, err := NewUserAuth("https://test.example.com", "user", "")
			So(c, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})
		Convey("should error with empty URL", func() {
			c, err := NewUserAuth("", "user", "password")
			So(c, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})
	})

	Convey("Valid NewUserAuth", t, func() {
		c, err := NewUserAuth("https://test.example.com", "user", "password")
		Convey("Should not error", func() {
			So(err, ShouldBeNil)
		})
		Convey("Should have a valid client", func() {
			So(c, ShouldNotBeNil)
		})
	})
}

func WithServer(status api.AuthStatus, returnCode int, token, expectedPath, expectedMethod string, expectedHeaders map[string]string, f func(ts *httptest.Server)) func() {
	return func() {
		Convey("http requests should be correct", func(c C) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				c.So(r.Method, ShouldEqual, expectedMethod)
				c.So(r.URL.Path, ShouldStartWith, expectedPath)
				// Check headers
				for k, v := range expectedHeaders {
					c.So(r.Header.Get(k), ShouldEqual, v)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(returnCode)
				w.Write([]byte(fmt.Sprintf(validLogin, status, token)))
			}))
			f(ts)
			Reset(func() {
				ts.Close()
			})
		})

	}
}

// This by extension will test the `authenticate` and `checkAndParse` methods
func TestGetToken(t *testing.T) {
	var token = "7f6808f1-ede3-2177-aa9d-45f507391310"
	Convey("GetToken with valid credentials", t, WithServer(api.AuthUserSuccess, http.StatusOK, token, "/v2/auth/user", http.MethodGet, map[string]string{}, func(ts *httptest.Server) {
		c, _ := NewUserAuth(ts.URL, "user", "password")
		So(c, ShouldNotBeNil)
		Convey("Should return a valid token", func() {
			t, err := c.GetToken(nil)
			So(err, ShouldBeNil)
			So(t, ShouldEqual, token)
			Convey("And should have a valid expiry time", func() {
				So(c.expiry, ShouldHappenOnOrBefore, time.Now().Add(3600*time.Second))
			})
			Convey("X-Vault-Token header should be set", func() {
				So(c.headers.Get("X-Vault-Token"), ShouldEqual, token)
			})
		})
	}))

	Convey("GetToken with invalid credentials", t, WithServer(api.AuthUserSuccess, http.StatusUnauthorized, token, "/v2/auth/user", http.MethodGet, map[string]string{}, func(ts *httptest.Server) {
		c, _ := NewUserAuth(ts.URL, "user", "password")
		So(c, ShouldNotBeNil)
		Convey("Should return an error", func() {
			t, err := c.GetToken(nil)
			So(err, ShouldNotBeNil)
			So(err, ShouldEqual, ErrorUnauthorized)
			So(t, ShouldBeEmpty)
		})
	}))

	Convey("GetToken with a bad request", t, WithServer(api.AuthUserSuccess, http.StatusBadRequest, token, "/v2/auth/user", http.MethodGet, map[string]string{}, func(ts *httptest.Server) {
		c, _ := NewUserAuth(ts.URL, "user", "password")
		So(c, ShouldNotBeNil)
		Convey("Should return an error", func() {
			t, err := c.GetToken(nil)
			So(err, ShouldNotBeNil)
			So(t, ShouldBeEmpty)
		})
	}))

	Convey("GetToken with a non responsive server", t, func() {
		c, _ := NewUserAuth("http://127.0.0.1:32876", "user", "password")
		So(c, ShouldNotBeNil)
		Convey("Should error", func() {
			_, err := c.GetToken(nil)
			So(err, ShouldNotBeNil)
		})
	})

	Convey("GetToken with valid credentials and MFA flow", t, func() {
		// We have to do our own special test server here to handle both requests to the API
		Convey("http requests should be correct", func(c C) {
			var firstRequest = true
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// First request assertions
				if firstRequest {
					c.So(r.Method, ShouldEqual, http.MethodGet)
					c.So(r.URL.Path, ShouldStartWith, "/v2/auth/user")
					firstRequest = false
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(fmt.Sprintf(validLoginMFA, api.AuthUserNeedsMFA, "5c7d1fd1914ffff5bcc2253b3c38ef85a3125bc1")))
					return
				}
				c.So(r.Method, ShouldEqual, http.MethodPost)
				c.So(r.URL.Path, ShouldStartWith, "/v2/auth/mfa_check")
				// Make sure the state token is there and that the otp_token is not empty
				body := map[string]string{}
				err := json.NewDecoder(r.Body).Decode(&body)
				c.So(err, ShouldBeNil)
				c.So(body, ShouldContainKey, "state_token")
				c.So(body["state_token"], ShouldEqual, "5c7d1fd1914ffff5bcc2253b3c38ef85a3125bc1")
				c.So(body, ShouldContainKey, "otp_token")
				c.So(body["otp_token"], ShouldNotBeEmpty)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(fmt.Sprintf(validLogin, api.AuthUserSuccess, token)))
			}))
			client, _ := NewUserAuth(ts.URL, "user", "password")
			So(client, ShouldNotBeNil)
			Convey("Should return a valid token", func() {
				// Create a temp file for testing the otp token
				in, err := ioutil.TempFile("", "")
				So(err, ShouldBeNil)
				defer in.Close()
				defer os.Remove(in.Name())
				io.WriteString(in, "acooltoken\n")
				// Reset the file to the beginning
				in.Seek(0, os.SEEK_SET)
				t, err := client.GetToken(in)
				So(err, ShouldBeNil)
				So(t, ShouldEqual, token)
				Convey("And should have a valid expiry time", func() {
					So(client.expiry, ShouldHappenOnOrBefore, time.Now().Add(3600*time.Second))
				})
			})
		})
	})
	Convey("GetToken if already authenticated", t, func() {
		c, _ := NewUserAuth("http://example.com", "user", "password")
		So(c, ShouldNotBeNil)
		c.setToken("test-token", 3600)
		Convey("Should return token", func() {
			t, err := c.GetToken(nil)
			So(err, ShouldBeNil)
			So(t, ShouldEqual, "test-token")
		})
	})
}

func TestRefresh(t *testing.T) {
	var token = "a-new-token"
	Convey("Refreshing a token", t, WithServer(api.AuthUserSuccess, http.StatusOK, token, "/v2/auth/user/refresh", http.MethodGet, map[string]string{"X-Vault-Token": "an-old-token"}, func(ts *httptest.Server) {
		c, _ := NewUserAuth(ts.URL, "user", "password")
		So(c, ShouldNotBeNil)
		c.setToken("an-old-token", 3600)
		Convey("Should return a new valid token", func() {
			err := c.Refresh()
			So(err, ShouldBeNil)
			So(c.token, ShouldEqual, token)
			Convey("And should have a valid expiry time", func() {
				So(c.expiry, ShouldHappenOnOrBefore, time.Now().Add(3600*time.Second))
			})
			Convey("X-Vault-Token header should be set", func() {
				So(c.headers.Get("X-Vault-Token"), ShouldEqual, token)
			})
		})
	}))

	Convey("Refreshing when not authenticated", t, func() {
		c, _ := NewUserAuth("http://example.com", "user", "password")
		So(c, ShouldNotBeNil)
		Convey("Should error", func() {
			err := c.Refresh()
			So(err, ShouldEqual, ErrorUnauthenticated)
		})
	})
	Convey("Refreshing with an expired token", t, func() {
		c, _ := NewUserAuth("http://example.com", "user", "password")
		So(c, ShouldNotBeNil)
		c.setToken("an-old-token", 3600)
		c.expiry = time.Now().Add(-2 * time.Minute)
		Convey("Should error", func() {
			err := c.Refresh()
			So(err, ShouldEqual, ErrorUnauthenticated)
		})
	})

	Convey("Refreshing with bad request", t, WithServer(api.AuthUserSuccess, http.StatusBadRequest, token, "/v2/auth/user/refresh", http.MethodGet, map[string]string{}, func(ts *httptest.Server) {
		c, _ := NewUserAuth(ts.URL, "user", "password")
		So(c, ShouldNotBeNil)
		c.setToken("an-old-token", 3600)
		Convey("Should error", func() {
			err := c.Refresh()
			So(err, ShouldNotBeNil)
		})
	}))
	// TODO: Figure out what the response looks like if you try to refresh with an invalid token.
	// Theoretically, this should never happen because the user can't mess with the token and this implementation
	// refreshes itself with username/password. But we probably should add a test case
}

func TestGetHeaders(t *testing.T) {
	Convey("Getting headers when not authenticated", t, func() {
		c, _ := NewUserAuth("http://example.com", "user", "password")
		So(c, ShouldNotBeNil)
		Convey("Should error", func() {
			_, err := c.GetHeaders()
			So(err, ShouldEqual, ErrorUnauthenticated)
		})
	})

	Convey("Getting headers with authenticated client", t, func() {
		c, _ := NewUserAuth("http://example.com", "user", "password")
		So(c, ShouldNotBeNil)
		c.setToken("an-old-token", 3600)
		headers, err := c.GetHeaders()
		Convey("Should not error", func() {
			So(err, ShouldBeNil)
		})
		Convey("Should contain X-Vault-Token", func() {
			So(headers.Get("X-Vault-Token"), ShouldEqual, "an-old-token")
		})
	})
}

func TestLogout(t *testing.T) {
	Convey("Logging out when not authenticated", t, func() {
		c, _ := NewUserAuth("http://example.com", "user", "password")
		So(c, ShouldNotBeNil)
		Convey("Should error", func() {
			err := c.Logout()
			So(err, ShouldEqual, ErrorUnauthenticated)
		})
	})

	Convey("Logging out with non-responsive server", t, func() {
		c, _ := NewUserAuth("http://127.0.0.1:32876", "user", "password")
		So(c, ShouldNotBeNil)
		Convey("Should error", func() {
			err := c.Logout()
			So(err, ShouldNotBeNil)
		})
	})

	Convey("Logging out with bad request", t, WithServer(api.AuthUserSuccess, http.StatusBadRequest, "", "/v1/auth", http.MethodDelete, map[string]string{}, func(ts *httptest.Server) {
		c, _ := NewUserAuth(ts.URL, "user", "password")
		So(c, ShouldNotBeNil)
		c.setToken("an-old-token", 3600)
		Convey("Should error", func() {
			err := c.Logout()
			So(err, ShouldNotBeNil)
		})
	}))

	Convey("Logging out with valid token", t, WithServer(api.AuthUserSuccess, http.StatusNoContent, "", "/v1/auth", http.MethodDelete, map[string]string{"X-Vault-Token": "an-old-token"}, func(ts *httptest.Server) {
		c, _ := NewUserAuth(ts.URL, "user", "password")
		So(c, ShouldNotBeNil)
		c.setToken("an-old-token", 3600)
		Convey("Should not error", func() {
			err := c.Logout()
			So(err, ShouldBeNil)
			Convey("Token should no longer be set", func() {
				So(c.token, ShouldBeEmpty)
			})
			Convey("Headers should no longer be set", func() {
				So(c.headers.Get("X-Vault-Token"), ShouldBeEmpty)
			})
		})
	}))
}
