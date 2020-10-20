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
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Nike-Inc/cerberus-go-client/v3/api"
	. "github.com/smartystreets/goconvey/convey"
)

var metadataTime, _ = time.Parse(time.RFC3339, "2017-01-04T23:18:40-08:00")

var metadataBody = `{
    "has_next": false,
    "next_offset": 0,
    "limit": 10,
    "offset": 0,
    "sdb_count_in_result": 2,
    "total_sdbcount": 2,
    "safe_deposit_box_metadata": [
        {
            "name": "dev demo",
            "path": "app/dev-demo/",
            "category": "Applications",
            "owner": "Lst-Squad.Carebears",
            "description": "test",
            "created_ts": "2017-01-04T23:18:40-08:00",
            "created_by": "justin.field@nike.com",
            "last_updated_ts": "2017-01-04T23:18:40-08:00",
            "last_updated_by": "justin.field@nike.com",
            "user_group_permissions": {
                "Application.FOO.User": "read"
            },
            "iam_role_permissions": {
                "arn:aws:iam::265866363820:role/asdf": "write"
            }
        },
        {
            "name": "IaM W d WASD",
            "path": "shared/iam-w-d-wasd/",
            "category": "Shared",
            "owner": "Lst-Squad.Carebears",
            "description": "CAREBERS",
            "created_ts": "2017-01-04T23:18:40-08:00",
            "created_by": "justin.field@nike.com",
            "last_updated_ts": "2017-01-04T23:18:40-08:00",
            "last_updated_by": "justin.field@nike.com",
            "user_group_permissions": {},
            "iam_role_permissions": {}
        }
    ]
}`

var expectedMetadata = &api.MetadataResponse{
	HasNext:     false,
	NextOffset:  0,
	Limit:       10,
	Offset:      0,
	ResultCount: 2,
	TotalCount:  2,
	Metadata: []api.SDBMetadata{
		api.SDBMetadata{
			Name:          "dev demo",
			Path:          "app/dev-demo/",
			Category:      "Applications",
			Owner:         "Lst-Squad.Carebears",
			Description:   "test",
			Created:       metadataTime,
			CreatedBy:     "justin.field@nike.com",
			LastUpdated:   metadataTime,
			LastUpdatedBy: "justin.field@nike.com",
			UserGroupPermissions: map[string]string{
				"Application.FOO.User": "read",
			},
			IAMRolePermissions: map[string]string{
				"arn:aws:iam::265866363820:role/asdf": "write",
			},
		},
		api.SDBMetadata{
			Name:                 "IaM W d WASD",
			Path:                 "shared/iam-w-d-wasd/",
			Category:             "Shared",
			Owner:                "Lst-Squad.Carebears",
			Description:          "CAREBERS",
			Created:              metadataTime,
			CreatedBy:            "justin.field@nike.com",
			LastUpdated:          metadataTime,
			LastUpdatedBy:        "justin.field@nike.com",
			UserGroupPermissions: map[string]string{},
			IAMRolePermissions:   map[string]string{},
		},
	},
}

func TestListMetadata(t *testing.T) {
	Convey("A valid call to List", t, WithTestServer(http.StatusOK, "/v1/metadata", http.MethodGet, metadataBody, func(ts *httptest.Server) {
		cl, _ := NewClient(GenerateMockAuth(ts.URL, "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		Convey("Should return a valid list of roles", func() {
			roles, err := cl.Metadata().List(MetadataOpts{})
			So(err, ShouldBeNil)
			So(roles, ShouldResemble, expectedMetadata)
		})
	}))

	Convey("An invalid call to List", t, WithTestServer(http.StatusInternalServerError, "/v1/metadata", http.MethodGet, "", func(ts *httptest.Server) {
		cl, _ := NewClient(GenerateMockAuth(ts.URL, "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		Convey("Should error", func() {
			roles, err := cl.Metadata().List(MetadataOpts{})
			So(err, ShouldNotBeNil)
			So(roles, ShouldBeNil)
		})
	}))

	Convey("Invalid params", t, WithTestServer(http.StatusBadRequest, "/v1/metadata", http.MethodGet, errorResponse, func(ts *httptest.Server) {
		cl, _ := NewClient(GenerateMockAuth(ts.URL, "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		Convey("Should error", func() {
			roles, err := cl.Metadata().List(MetadataOpts{Offset: 1000000})
			So(err, ShouldNotBeNil)
			So(roles, ShouldBeNil)
			Convey("And return an API ErrorResponse", func() {
				So(err, ShouldResemble, expectedError)
			})
		})
	}))

	Convey("A List to a non-responsive server", t, func() {
		cl, _ := NewClient(GenerateMockAuth("http://127.0.0.1:32876", "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		Convey("Should return an error", func() {
			roles, err := cl.Metadata().List(MetadataOpts{})
			So(err, ShouldNotBeNil)
			So(roles, ShouldBeNil)
		})
	})
}
