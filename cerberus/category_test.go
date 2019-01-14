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

	"github.com/Nike-Inc/cerberus-go-client/api"
	. "github.com/smartystreets/goconvey/convey"
)

var categoryResponse = `[
    {
        "id": "f7ff85a0-faaa-11e5-a8a9-7fa3b294cd46",
        "display_name": "Applications",
        "path": "app",
        "created_ts": "2016-04-05T04:19:51Z",
        "last_updated_ts": "2016-04-05T04:19:51Z",
        "created_by": "system",
        "last_updated_by": "system"
    },
    {
        "id": "f7ffb890-faaa-11e5-a8a9-7fa3b294cd46",
        "display_name": "Shared",
        "path": "shared",
        "created_ts": "2016-04-05T04:19:51Z",
        "last_updated_ts": "2016-04-05T04:19:51Z",
        "created_by": "system",
        "last_updated_by": "system"
    }
]`

var listTime, _ = time.Parse(time.RFC3339, "2016-04-05T04:19:51Z")

var expectedResponseList = []*api.Category{
	&api.Category{
		ID:            "f7ff85a0-faaa-11e5-a8a9-7fa3b294cd46",
		DisplayName:   "Applications",
		Path:          "app",
		Created:       listTime,
		LastUpdated:   listTime,
		CreatedBy:     "system",
		LastUpdatedBy: "system",
	},
	&api.Category{
		ID:            "f7ffb890-faaa-11e5-a8a9-7fa3b294cd46",
		DisplayName:   "Shared",
		Path:          "shared",
		Created:       listTime,
		LastUpdated:   listTime,
		CreatedBy:     "system",
		LastUpdatedBy: "system",
	},
}

func TestListCategory(t *testing.T) {
	Convey("A valid call to List", t, WithTestServer(http.StatusOK, "/v1/category", http.MethodGet, categoryResponse, func(ts *httptest.Server) {
		cl, _ := NewClient(GenerateMockAuth(ts.URL, "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		Convey("Should return a valid list of categories", func() {
			categories, err := cl.Category().List()
			So(err, ShouldBeNil)
			So(categories, ShouldResemble, expectedResponseList)
		})
	}))

	Convey("An invalid call to List", t, WithTestServer(http.StatusInternalServerError, "/v1/category", http.MethodGet, "", func(ts *httptest.Server) {
		cl, _ := NewClient(GenerateMockAuth(ts.URL, "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		Convey("Should error", func() {
			categories, err := cl.Category().List()
			So(err, ShouldNotBeNil)
			So(categories, ShouldBeNil)
		})
	}))

	Convey("A List to a non-responsive server", t, func() {
		cl, _ := NewClient(GenerateMockAuth("http://127.0.0.1:32876", "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		Convey("Should return an error", func() {
			categories, err := cl.Category().List()
			So(err, ShouldNotBeNil)
			So(categories, ShouldBeNil)
		})
	})
}
