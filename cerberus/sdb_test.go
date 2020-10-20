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

package cerberus

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Nike-Inc/cerberus-go-client/v3/api"
	. "github.com/smartystreets/goconvey/convey"
)

func WithTestServer(returnCode int, expectedPath, expectedMethod, body string, f func(ts *httptest.Server)) func() {
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

var errorResponse = `{
	"error_id": "a041aa4d-1d5a-4eed-8e8a-6dc18bdf96db",
	"errors": [{
		"code": 99208,
		"message": "The name may not be blank.",
		"metadata": {
			"field": "name"
		}
	}]
}`

var expectedError = api.ErrorResponse{
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

func TestGetSDB(t *testing.T) {
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
    "iam_principal_permissions": [
        {
            "id": "d05bf72e-faad-11e5-a8a9-7fa3b294cd46",
            "iam_principal_arn": "arn:aws:iam::1111111111:role/role-name",
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
		IAMPrincipalPermissions: []api.IAMPrincipal{
			api.IAMPrincipal{
				ID:              "d05bf72e-faad-11e5-a8a9-7fa3b294cd46",
				IAMPrincipalARN: "arn:aws:iam::1111111111:role/role-name",
				RoleID:          "f800558e-faaa-11e5-a8a9-7fa3b294cd46",
			},
		},
	}

	Convey("A valid GET of ID", t, WithTestServer(http.StatusOK, fmt.Sprintf("/v2/safe-deposit-box/%s", id), http.MethodGet, fmt.Sprintf(validResponse, id), func(ts *httptest.Server) {
		cl, _ := NewClient(GenerateMockAuth(ts.URL, "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		Convey("Should return a valid SDB", func() {
			box, err := cl.SDB().Get(id)
			So(err, ShouldBeNil)
			So(box, ShouldResemble, expectedResponse)
		})
	}))

	Convey("A GET of nonexistent ID", t, WithTestServer(http.StatusNotFound, fmt.Sprintf("/v2/safe-deposit-box/%s", id), http.MethodGet, "", func(ts *httptest.Server) {
		cl, _ := NewClient(GenerateMockAuth(ts.URL, "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		Convey("Should return SDB not found error", func() {
			box, err := cl.SDB().Get(id)
			So(err, ShouldEqual, ErrorSafeDepositBoxNotFound)
			So(box, ShouldBeNil)
		})
	}))

	Convey("A GET request that encounters a server error", t, WithTestServer(http.StatusInternalServerError, fmt.Sprintf("/v2/safe-deposit-box/%s", id), http.MethodGet, "", func(ts *httptest.Server) {
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

func TestListSDB(t *testing.T) {
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

	Convey("A valid call to List", t, WithTestServer(http.StatusOK, "/v2/safe-deposit-box", http.MethodGet, validResponse, func(ts *httptest.Server) {
		cl, _ := NewClient(GenerateMockAuth(ts.URL, "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		Convey("Should return a valid list of SDB", func() {
			boxes, err := cl.SDB().List()
			So(err, ShouldBeNil)
			So(boxes, ShouldResemble, expectedResponse)
		})
	}))

	Convey("A call to List that encounters a server error", t, WithTestServer(http.StatusInternalServerError, "/v2/safe-deposit-box", http.MethodGet, validResponse, func(ts *httptest.Server) {
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

	Convey("A valid call to GetByName", t, WithTestServer(http.StatusOK, "/v2/safe-deposit-box", http.MethodGet, validResponse, func(ts *httptest.Server) {
		cl, _ := NewClient(GenerateMockAuth(ts.URL, "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		Convey("Should return a valid SDB", func() {
			box, err := cl.SDB().GetByName("Web")
			So(err, ShouldBeNil)
			So(box, ShouldResemble, expectedResponse)
		})
	}))

	Convey("GetByName given an invalid name", t, WithTestServer(http.StatusOK, "/v2/safe-deposit-box", http.MethodGet, validResponse, func(ts *httptest.Server) {
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

	Convey("A call to GetByName that encounters a server error", t, WithTestServer(http.StatusInternalServerError, "/v2/safe-deposit-box", http.MethodGet, validResponse, func(ts *httptest.Server) {
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

func TestCreateSDB(t *testing.T) {
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
    "iam_principal_permissions": [
        {
            "id": "d05bf72e-faad-11e5-a8a9-7fa3b294cd46",
            "iam_principal_arn": "arn:aws:iam::1111111111:role/role-name",
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
		IAMPrincipalPermissions: []api.IAMPrincipal{
			api.IAMPrincipal{
				ID:              "d05bf72e-faad-11e5-a8a9-7fa3b294cd46",
				IAMPrincipalARN: "arn:aws:iam::1111111111:role/role-name",
				RoleID:          "f800558e-faaa-11e5-a8a9-7fa3b294cd46",
			},
		},
	}

	var newSDB = &api.SafeDepositBox{
		Name:        "Stage",
		Description: "Sensitive configuration properties for the stage micro-service.",
		CategoryID:  "f7ff85a0-faaa-11e5-a8a9-7fa3b294cd46",
		Owner:       "Lst-digital.platform-tools.internal",
		UserGroupPermissions: []api.UserGroupPermission{
			api.UserGroupPermission{
				ID:     "3fc6455c-faad-11e5-a8a9-7fa3b294cd46",
				Name:   "Lst-CDT.CloudPlatformEngine.FTE",
				RoleID: "f800558e-faaa-11e5-a8a9-7fa3b294cd46",
			},
		},
		IAMPrincipalPermissions: []api.IAMPrincipal{
			api.IAMPrincipal{
				ID:              "d05bf72e-faad-11e5-a8a9-7fa3b294cd46",
				IAMPrincipalARN: "arn:aws:iam::1111111111:role/role-name",
				RoleID:          "f800558e-faaa-11e5-a8a9-7fa3b294cd46",
			},
		},
	}

	Convey("A valid new SDB object", t, WithTestServer(http.StatusCreated, "/v2/safe-deposit-box", http.MethodPost, fmt.Sprintf(validResponse, id), func(ts *httptest.Server) {
		cl, _ := NewClient(GenerateMockAuth(ts.URL, "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		Convey("Should create successfully", func() {
			box, err := cl.SDB().Create(newSDB)
			So(err, ShouldBeNil)
			Convey("And return a valid object", func() {
				So(box, ShouldResemble, expectedResponse)
			})
		})
	}))

	Convey("An invalid new SDB object", t, WithTestServer(http.StatusBadRequest, "/v2/safe-deposit-box", http.MethodPost, errorResponse, func(ts *httptest.Server) {
		cl, _ := NewClient(GenerateMockAuth(ts.URL, "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		var badSDB = *newSDB
		badSDB.Name = ""
		Convey("Should error", func() {
			box, err := cl.SDB().Create(&badSDB)
			So(err, ShouldNotBeNil)
			So(box, ShouldBeNil)
			Convey("And return an API ErrorResponse", func() {
				So(err, ShouldResemble, expectedError)
			})
		})
	}))

	Convey("An bad server response", t, WithTestServer(http.StatusInternalServerError, "/v2/safe-deposit-box", http.MethodPost, "", func(ts *httptest.Server) {
		cl, _ := NewClient(GenerateMockAuth(ts.URL, "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		Convey("Should error", func() {
			box, err := cl.SDB().Create(newSDB)
			So(err, ShouldNotBeNil)
			So(box, ShouldBeNil)
			Convey("And should not be an API ErrorResponse", func() {
				So(err, ShouldNotHaveSameTypeAs, api.ErrorResponse{})
			})
		})
	}))

	Convey("An bad server response with an invalid body", t, WithTestServer(http.StatusInternalServerError, "/v2/safe-deposit-box", http.MethodPost, "blah", func(ts *httptest.Server) {
		cl, _ := NewClient(GenerateMockAuth(ts.URL, "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		Convey("Should error", func() {
			box, err := cl.SDB().Create(newSDB)
			So(err, ShouldNotBeNil)
			So(box, ShouldBeNil)
			Convey("And should not be an API ErrorResponse", func() {
				So(err, ShouldNotHaveSameTypeAs, api.ErrorResponse{})
			})
		})
	}))

	Convey("A Create to a non-responsive server", t, func() {
		cl, _ := NewClient(GenerateMockAuth("http://127.0.0.1:32876", "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		Convey("Should return an error", func() {
			box, err := cl.SDB().Create(newSDB)
			So(err, ShouldNotBeNil)
			So(box, ShouldBeNil)
		})
	})
}

func TestUpdateSDB(t *testing.T) {
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
    "iam_principal_permissions": [
        {
            "id": "d05bf72e-faad-11e5-a8a9-7fa3b294cd46",
            "iam_principal_arn": "arn:aws:iam::1111111111:role/role-name",
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
		IAMPrincipalPermissions: []api.IAMPrincipal{
			api.IAMPrincipal{
				ID:              "d05bf72e-faad-11e5-a8a9-7fa3b294cd46",
				IAMPrincipalARN: "arn:aws:iam::1111111111:role/role-name",
				RoleID:          "f800558e-faaa-11e5-a8a9-7fa3b294cd46",
			},
		},
	}

	var updated = &api.SafeDepositBox{
		Description: "Sensitive configuration properties for the stage micro-service.",
		Owner:       "Lst-digital.platform-tools.internal",
	}

	Convey("A valid SDB object", t, WithTestServer(http.StatusOK, "/v2/safe-deposit-box/"+id, http.MethodPut, fmt.Sprintf(validResponse, id), func(ts *httptest.Server) {
		cl, _ := NewClient(GenerateMockAuth(ts.URL, "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		Convey("Should update successfully", func() {
			box, err := cl.SDB().Update(id, updated)
			So(err, ShouldBeNil)
			Convey("And return a valid object", func() {
				So(box, ShouldResemble, expectedResponse)
			})
		})
	}))

	Convey("An invalid SDB object", t, WithTestServer(http.StatusBadRequest, "/v2/safe-deposit-box/"+id, http.MethodPut, errorResponse, func(ts *httptest.Server) {
		cl, _ := NewClient(GenerateMockAuth(ts.URL, "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		var badSDB = *updated
		badSDB.ID = "you-shouldn't-change-this"
		Convey("Should error", func() {
			box, err := cl.SDB().Update(id, &badSDB)
			So(err, ShouldNotBeNil)
			So(box, ShouldBeNil)
			Convey("And return an API ErrorResponse", func() {
				So(err, ShouldResemble, expectedError)
			})
		})
	}))

	Convey("An update to a non-existent ID", t, WithTestServer(http.StatusNotFound, "/v2/safe-deposit-box/blah", http.MethodPut, "blah", func(ts *httptest.Server) {
		cl, _ := NewClient(GenerateMockAuth(ts.URL, "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		Convey("Should error", func() {
			box, err := cl.SDB().Update("blah", updated)
			So(err, ShouldNotBeNil)
			So(box, ShouldBeNil)
			Convey("And should be the right error type", func() {
				So(err, ShouldEqual, ErrorSafeDepositBoxNotFound)
			})
		})
	}))

	Convey("An bad server response", t, WithTestServer(http.StatusInternalServerError, "/v2/safe-deposit-box/"+id, http.MethodPut, "", func(ts *httptest.Server) {
		cl, _ := NewClient(GenerateMockAuth(ts.URL, "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		Convey("Should error", func() {
			box, err := cl.SDB().Update(id, updated)
			So(err, ShouldNotBeNil)
			So(box, ShouldBeNil)
			Convey("And should not be an API ErrorResponse", func() {
				So(err, ShouldNotHaveSameTypeAs, api.ErrorResponse{})
			})
		})
	}))

	Convey("An bad server response with an invalid body", t, WithTestServer(http.StatusInternalServerError, "/v2/safe-deposit-box/"+id, http.MethodPut, "blah", func(ts *httptest.Server) {
		cl, _ := NewClient(GenerateMockAuth(ts.URL, "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		Convey("Should error", func() {
			box, err := cl.SDB().Update(id, updated)
			So(err, ShouldNotBeNil)
			So(box, ShouldBeNil)
			Convey("And should not be an API ErrorResponse", func() {
				So(err, ShouldNotHaveSameTypeAs, api.ErrorResponse{})
			})
		})
	}))

	Convey("An Update to a non-responsive server", t, func() {
		cl, _ := NewClient(GenerateMockAuth("http://127.0.0.1:32876", "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		Convey("Should return an error", func() {
			box, err := cl.SDB().Update(id, updated)
			So(err, ShouldNotBeNil)
			So(box, ShouldBeNil)
		})
	})

	Convey("A call to Update with an empty ID", t, func() {
		cl, _ := NewClient(GenerateMockAuth("http://127.0.0.1:32876", "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		Convey("Should return an error", func() {
			box, err := cl.SDB().Update("", updated)
			So(err, ShouldNotBeNil)
			So(err, ShouldEqual, ErrorSafeDepositBoxNotFound)
			So(box, ShouldBeNil)
		})
	})
}

func TestDeleteSDB(t *testing.T) {
	var id = "a7d703da-faac-11e5-a8a9-7fa3b294cd46"

	Convey("A valid delete", t, WithTestServer(http.StatusNoContent, "/v2/safe-deposit-box/"+id, http.MethodDelete, "", func(ts *httptest.Server) {
		cl, _ := NewClient(GenerateMockAuth(ts.URL, "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		Convey("Should complete successfully", func() {
			err := cl.SDB().Delete(id)
			So(err, ShouldBeNil)
		})
	}))

	Convey("An invalid delete", t, WithTestServer(http.StatusBadRequest, "/v2/safe-deposit-box/"+id, http.MethodDelete, errorResponse, func(ts *httptest.Server) {
		cl, _ := NewClient(GenerateMockAuth(ts.URL, "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		Convey("Should error", func() {
			err := cl.SDB().Delete(id)
			So(err, ShouldNotBeNil)
			Convey("And return an API ErrorResponse", func() {
				So(err, ShouldResemble, expectedError)
			})
		})
	}))

	Convey("An delete of a non-existent ID", t, WithTestServer(http.StatusNotFound, "/v2/safe-deposit-box/blah", http.MethodDelete, "blah", func(ts *httptest.Server) {
		cl, _ := NewClient(GenerateMockAuth(ts.URL, "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		Convey("Should error", func() {
			err := cl.SDB().Delete("blah")
			So(err, ShouldNotBeNil)
			Convey("And should be the right error type", func() {
				So(err, ShouldEqual, ErrorSafeDepositBoxNotFound)
			})
		})
	}))

	Convey("An bad server response", t, WithTestServer(http.StatusInternalServerError, "/v2/safe-deposit-box/"+id, http.MethodDelete, "", func(ts *httptest.Server) {
		cl, _ := NewClient(GenerateMockAuth(ts.URL, "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		Convey("Should error", func() {
			err := cl.SDB().Delete(id)
			So(err, ShouldNotBeNil)
			Convey("And should not be an API ErrorResponse", func() {
				So(err, ShouldNotHaveSameTypeAs, api.ErrorResponse{})
			})
		})
	}))

	Convey("A delete to a non-responsive server", t, func() {
		cl, _ := NewClient(GenerateMockAuth("http://127.0.0.1:32876", "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		Convey("Should return an error", func() {
			err := cl.SDB().Delete(id)
			So(err, ShouldNotBeNil)
		})
	})

	Convey("A call to Delete with an empty ID", t, func() {
		cl, _ := NewClient(GenerateMockAuth("http://127.0.0.1:32876", "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		Convey("Should return an error", func() {
			err := cl.SDB().Delete("")
			So(err, ShouldNotBeNil)
			So(err, ShouldEqual, ErrorSafeDepositBoxNotFound)
		})
	})

}
