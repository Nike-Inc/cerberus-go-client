package cerberus

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"fmt"

	. "github.com/smartystreets/goconvey/convey"
	"github.nike.com/ngp/cerberus-client-go/api"
)

func SDBServer(returnCode int, expectedPath, expectedMethod, body string, f func(ts *httptest.Server)) func() {
	return func() {
		Convey("http requests should be correct", func(c C) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				c.So(r.Method, ShouldEqual, expectedMethod)
				c.So(r.URL.Path, ShouldStartWith, expectedPath)
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

func TestGet(t *testing.T) {
	var id = "a7d703da-faac-11e5-a8a9-7fa3b294cd46"
	var validResponse = `{
		"id": "%s",
		"name": "Stage",
		"description": "Sensitive configuration properties for the stage micro-service.",
		"path": "app/stage",
		"category_id": "f7ff85a0-faaa-11e5-a8a9-7fa3b294cd46",
		"owner": "Lst-digital.platform-tools.internal",
		"user_group_permissions": [
			{
				"id": "3fc6455c-faad-11e5-a8a9-7fa3b294cd46",
				"name": "Lst-CDT.CloudPlatformEngine.FTE",
				"role_id": "f800558e-faaa-11e5-a8a9-7fa3b294cd46"
			}
		],
		"iam_role_permissions": [
			{
				"id": "d05bf72e-faad-11e5-a8a9-7fa3b294cd46",
				"account_id": "123",
				"iam_role_name": "stage",
				"role_id": "f800558e-faaa-11e5-a8a9-7fa3b294cd46"
			}
		]
	}`

	var expectedResponse = &api.SafeDepositBox{
		ID:          id,
		Name:        "Stage",
		Description: "Sensitive configuration properties for the stage micro-service.",
		Path:        "app/stage",
		CategoryID:  "f7ff85a0-faaa-11e5-a8a9-7fa3b294cd46",
		Owner:       "Lst-digital.platform-tools.internal",
		UserGroupPermissions: []api.UserGroupPermission{
			api.UserGroupPermission{
				ID:     "3fc6455c-faad-11e5-a8a9-7fa3b294cd46",
				Name:   "Lst-CDT.CloudPlatformEngine.FTE",
				RoleID: "f800558e-faaa-11e5-a8a9-7fa3b294cd46",
			},
		},
		IAMRolePermissions: []api.IAMRole{
			api.IAMRole{
				ID:        "d05bf72e-faad-11e5-a8a9-7fa3b294cd46",
				AccountID: "123",
				Name:      "stage",
				RoleID:    "f800558e-faaa-11e5-a8a9-7fa3b294cd46",
			},
		},
	}

	Convey("A valid GET of ID", t, SDBServer(http.StatusOK, fmt.Sprintf("/v1/safe-deposit-box/%s", id), http.MethodGet, fmt.Sprintf(validResponse, id), func(ts *httptest.Server) {
		cl, _ := NewClient(GenerateMockAuth(ts.URL, "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		Convey("Should return a valid SDB", func() {
			box, err := cl.SDB().Get(id)
			So(err, ShouldBeNil)
			So(box, ShouldResemble, expectedResponse)
		})
	}))

	Convey("A GET of nonexistent ID", t, SDBServer(http.StatusNotFound, fmt.Sprintf("/v1/safe-deposit-box/%s", id), http.MethodGet, "", func(ts *httptest.Server) {
		cl, _ := NewClient(GenerateMockAuth(ts.URL, "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		Convey("Should return SDB not found error", func() {
			box, err := cl.SDB().Get(id)
			So(err, ShouldEqual, ErrorSafeDepositBoxNotFound)
			So(box, ShouldBeNil)
		})
	}))

	Convey("A GET request that encounters a server error", t, SDBServer(http.StatusInternalServerError, fmt.Sprintf("/v1/safe-deposit-box/%s", id), http.MethodGet, "", func(ts *httptest.Server) {
		cl, _ := NewClient(GenerateMockAuth(ts.URL, "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		Convey("Should return an error", func() {
			box, err := cl.SDB().Get(id)
			So(err, ShouldNotBeNil)
			So(box, ShouldBeNil)
		})
	}))

	Convey("A GET to a non-responsive server", t, func() {
		cl, _ := NewClient(GenerateMockAuth("http://127.0.0.1:32876", "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		Convey("Should return an error", func() {
			box, err := cl.SDB().Get(id)
			So(err, ShouldNotBeNil)
			So(box, ShouldBeNil)
		})
	})

	Convey("A GET with an empty ID", t, func() {
		cl, _ := NewClient(GenerateMockAuth("http://127.0.0.1:32876", "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		Convey("Should return an error", func() {
			box, err := cl.SDB().Get("")
			So(err, ShouldNotBeNil)
			So(box, ShouldBeNil)
		})
	})
}

func TestList(t *testing.T) {
	var validResponse = `[
		{
			"id": "fb013540-fb5f-11e5-ba72-e899458df21a",
			"name": "Web",
			"path": "app/web",
			"category_id": "f7ff85a0-faaa-11e5-a8a9-7fa3b294cd46"
		},
		{
			"id": "06f82494-fb60-11e5-ba72-e899458df21a",
			"name": "OneLogin",
			"path": "shared/onelogin",
			"category_id": "f7ffb890-faaa-11e5-a8a9-7fa3b294cd46"
		}
	]`

	var expectedResponse = []*api.SafeDepositBox{
		&api.SafeDepositBox{
			ID:         "fb013540-fb5f-11e5-ba72-e899458df21a",
			Name:       "Web",
			Path:       "app/web",
			CategoryID: "f7ff85a0-faaa-11e5-a8a9-7fa3b294cd46",
		},
		&api.SafeDepositBox{
			ID:         "06f82494-fb60-11e5-ba72-e899458df21a",
			Name:       "OneLogin",
			Path:       "shared/onelogin",
			CategoryID: "f7ffb890-faaa-11e5-a8a9-7fa3b294cd46",
		},
	}

	Convey("A valid call to List", t, SDBServer(http.StatusOK, "/v1/safe-deposit-box", http.MethodGet, validResponse, func(ts *httptest.Server) {
		cl, _ := NewClient(GenerateMockAuth(ts.URL, "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		Convey("Should return a valid list of SDB", func() {
			boxes, err := cl.SDB().List()
			So(err, ShouldBeNil)
			So(boxes, ShouldResemble, expectedResponse)
		})
	}))

	Convey("A call to List that encounters a server error", t, SDBServer(http.StatusInternalServerError, "/v1/safe-deposit-box", http.MethodGet, validResponse, func(ts *httptest.Server) {
		cl, _ := NewClient(GenerateMockAuth(ts.URL, "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		Convey("Should return an error", func() {
			boxes, err := cl.SDB().List()
			So(err, ShouldNotBeNil)
			So(boxes, ShouldBeNil)
		})
	}))

	Convey("A List to a non-responsive server", t, func() {
		cl, _ := NewClient(GenerateMockAuth("http://127.0.0.1:32876", "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		Convey("Should return an error", func() {
			boxes, err := cl.SDB().List()
			So(err, ShouldNotBeNil)
			So(boxes, ShouldBeNil)
		})
	})
}

func TestGetByName(t *testing.T) {
	var validResponse = `[
		{
			"id": "fb013540-fb5f-11e5-ba72-e899458df21a",
			"name": "Web",
			"path": "app/web",
			"category_id": "f7ff85a0-faaa-11e5-a8a9-7fa3b294cd46"
		},
		{
			"id": "06f82494-fb60-11e5-ba72-e899458df21a",
			"name": "OneLogin",
			"path": "shared/onelogin",
			"category_id": "f7ffb890-faaa-11e5-a8a9-7fa3b294cd46"
		}
	]`

	var expectedResponse = &api.SafeDepositBox{
		ID:         "fb013540-fb5f-11e5-ba72-e899458df21a",
		Name:       "Web",
		Path:       "app/web",
		CategoryID: "f7ff85a0-faaa-11e5-a8a9-7fa3b294cd46",
	}

	Convey("A valid call to GetByName", t, SDBServer(http.StatusOK, "/v1/safe-deposit-box", http.MethodGet, validResponse, func(ts *httptest.Server) {
		cl, _ := NewClient(GenerateMockAuth(ts.URL, "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		Convey("Should return a valid SDB", func() {
			box, err := cl.SDB().GetByName("Web")
			So(err, ShouldBeNil)
			So(box, ShouldResemble, expectedResponse)
		})
	}))

	Convey("GetByName given an invalid name", t, SDBServer(http.StatusOK, "/v1/safe-deposit-box", http.MethodGet, validResponse, func(ts *httptest.Server) {
		cl, _ := NewClient(GenerateMockAuth(ts.URL, "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		Convey("Should return SDB not found error", func() {
			box, err := cl.SDB().GetByName("Blah")
			So(err, ShouldEqual, ErrorSafeDepositBoxNotFound)
			So(box, ShouldBeNil)
		})
	}))

	Convey("A call to GetByName with an empty name", t, func() {
		cl, _ := NewClient(GenerateMockAuth("http://127.0.0.1:32876", "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		Convey("Should return an error", func() {
			box, err := cl.SDB().GetByName("")
			So(err, ShouldNotBeNil)
			So(box, ShouldBeNil)
		})
	})

	Convey("A call to GetByName that encounters a server error", t, SDBServer(http.StatusInternalServerError, "/v1/safe-deposit-box", http.MethodGet, validResponse, func(ts *httptest.Server) {
		cl, _ := NewClient(GenerateMockAuth(ts.URL, "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		Convey("Should return an error", func() {
			box, err := cl.SDB().GetByName("Web")
			So(err, ShouldNotBeNil)
			So(box, ShouldBeNil)
		})
	}))

	Convey("A GetByName to a non-responsive server", t, func() {
		cl, _ := NewClient(GenerateMockAuth("http://127.0.0.1:32876", "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		Convey("Should return an error", func() {
			box, err := cl.SDB().GetByName("Web")
			So(err, ShouldNotBeNil)
			So(box, ShouldBeNil)
		})
	})
}
