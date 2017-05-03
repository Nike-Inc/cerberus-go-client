package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/Nike-Inc/cerberus-go-client/api"
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
