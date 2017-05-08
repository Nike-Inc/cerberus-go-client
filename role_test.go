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

	. "github.com/smartystreets/goconvey/convey"
	"github.com/Nike-Inc/cerberus-go-client/api"
)

var listResponse = `[
    {
        "id": "f7fff4d6-faaa-11e5-a8a9-7fa3b294cd46",
        "name": "owner",
        "created_ts": "2016-04-05T04:19:51Z",
        "last_updated_ts": "2016-04-05T04:19:51Z",
        "created_by": "system",
        "last_updated_by": "system"
    },
    {
        "id": "f800558e-faaa-11e5-a8a9-7fa3b294cd46",
        "name": "read",
        "created_ts": "2016-04-05T04:19:51Z",
        "last_updated_ts": "2016-04-05T04:19:51Z",
        "created_by": "system",
        "last_updated_by": "system"
    }
]`

var parsedTime, _ = time.Parse(time.RFC3339, "2016-04-05T04:19:51Z")

var expectedList = []*api.Role{
	&api.Role{
		ID:            "f7fff4d6-faaa-11e5-a8a9-7fa3b294cd46",
		Name:          "owner",
		Created:       parsedTime,
		LastUpdated:   parsedTime,
		CreatedBy:     "system",
		LastUpdatedBy: "system",
	},
	&api.Role{
		ID:            "f800558e-faaa-11e5-a8a9-7fa3b294cd46",
		Name:          "read",
		Created:       parsedTime,
		LastUpdated:   parsedTime,
		CreatedBy:     "system",
		LastUpdatedBy: "system",
	},
}

func TestListRole(t *testing.T) {
	Convey("A valid call to List", t, WithTestServer(http.StatusOK, "/v1/role", http.MethodGet, listResponse, func(ts *httptest.Server) {
		cl, _ := NewClient(GenerateMockAuth(ts.URL, "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		Convey("Should return a valid list of roles", func() {
			roles, err := cl.Role().List()
			So(err, ShouldBeNil)
			So(roles, ShouldResemble, expectedList)
		})
	}))

	Convey("An invalid call to List", t, WithTestServer(http.StatusInternalServerError, "/v1/role", http.MethodGet, "", func(ts *httptest.Server) {
		cl, _ := NewClient(GenerateMockAuth(ts.URL, "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		Convey("Should error", func() {
			roles, err := cl.Role().List()
			So(err, ShouldNotBeNil)
			So(roles, ShouldBeNil)
		})
	}))

	Convey("A List to a non-responsive server", t, func() {
		cl, _ := NewClient(GenerateMockAuth("http://127.0.0.1:32876", "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		Convey("Should return an error", func() {
			roles, err := cl.Role().List()
			So(err, ShouldNotBeNil)
			So(roles, ShouldBeNil)
		})
	})
}
