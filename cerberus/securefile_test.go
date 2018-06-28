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
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Nike-Inc/cerberus-go-client/api"
	. "github.com/smartystreets/goconvey/convey"
)

var secureFileListReply = `{
	"has_next" : false,
	"next_offset" : null,
	"limit" : 1000,
	"offset" : 0,
	"file_count_in_result" : 1,
	"total_file_count" : 1,
	"secure_file_summaries" : [ {
	  "sdbox_id" : "3f40b0ca-f7e4-4e38-bf1f-c36e05e1856f",
	  "path" : "godmiljaar/README.md",
	  "size_in_bytes" : 3296,
	  "name" : "README.md",
	  "created_by" : "pierre.poissinger@nike.com",
	  "created_ts" : "2018-06-14T10:34:55.057Z",
	  "last_updated_by" : "pierre.poissinger@nike.com",
	  "last_updated_ts" : "2018-06-14T10:34:55.057Z"
	} ]
  }`

var secureFileListReplyTime, _ = time.Parse(time.RFC3339, "2018-06-14T10:34:55.057Z")

var expectedResponse = &api.SecureFilesResponse{
	HasNext:     false,
	NextOffset:  0,
	Limit:       1000,
	Offset:      0,
	ResultCount: 1,
	TotalCount:  1,
	Summaries: []api.SecureFileSummary{
		{
			SDBID:         "3f40b0ca-f7e4-4e38-bf1f-c36e05e1856f",
			Path:          "godmiljaar/README.md",
			Size:          3296,
			Name:          "README.md",
			CreatedBy:     "pierre.poissinger@nike.com",
			Created:       secureFileListReplyTime,
			LastUpdatedBy: "pierre.poissinger@nike.com",
			LastUpdated:   secureFileListReplyTime,
		},
	},
}

func withBinaryTestServer(returnCode int, expectedPath, expectedMethod, filename string, body []byte, f func(ts *httptest.Server)) func() {
	return func() {
		Convey("http requests should be correct", func(c C) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				c.So(r.Method, ShouldEqual, expectedMethod)
				c.So(r.URL.Path, ShouldStartWith, expectedPath)
				w.Header().Set("Content-Type", "application/octet-stream")
				w.Header().Set("Content-Disposition",
					fmt.Sprintf(`attachment; filename="%s"`,
						filename))

				w.WriteHeader(returnCode)
				w.Write(body)
			}))
			f(ts)
			Reset(func() {
				ts.Close()
			})
		})

	}
}

func TestSecureFileList(t *testing.T) {
	Convey("A valid call to List", t, WithTestServer(http.StatusOK, "/v1/secure-files/my/sdb", http.MethodGet, secureFileListReply, func(ts *httptest.Server) {
		cl, _ := NewClient(GenerateMockAuth(ts.URL, "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		Convey("Should return a valid list of categories", func() {
			files, err := cl.SecureFile().List("my/sdb")
			So(err, ShouldBeNil)
			So(files, ShouldResemble, expectedResponse)
		})
	}))

	Convey("An invalid call to List", t, WithTestServer(http.StatusInternalServerError, "/v1/secure-files/my/sdb", http.MethodGet, "", func(ts *httptest.Server) {
		cl, _ := NewClient(GenerateMockAuth(ts.URL, "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		Convey("Should error", func() {
			files, err := cl.SecureFile().List("my/sdb")
			So(err, ShouldNotBeNil)
			So(files, ShouldBeNil)
		})
	}))

	Convey("A List to a non-responsive server", t, func() {
		cl, _ := NewClient(GenerateMockAuth("http://127.0.0.1:32876", "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		Convey("Should return an error", func() {
			files, err := cl.SecureFile().List("my/sdb")
			So(err, ShouldNotBeNil)
			So(files, ShouldBeNil)
		})
	})
}

func TestSecureFileGet(t *testing.T) {
	var fileBuffer bytes.Buffer

	Convey("A valid call to download", t, withBinaryTestServer(http.StatusOK,
		"/v1/secure-file/test/file/hello.txt",
		http.MethodGet,
		"hello.txt",
		[]byte("hello world"),
		func(ts *httptest.Server) {
			cl, _ := NewClient(GenerateMockAuth(ts.URL, "a-cool-token", false, false), nil)
			So(cl, ShouldNotBeNil)
			Convey("Should return a valid file", func() {
				fileBuffer.Reset()
				err := cl.SecureFile().Get("/test/file/hello.txt", &fileBuffer)
				So(err, ShouldBeNil)
				So(fileBuffer.Bytes(), ShouldResemble, []byte("hello world"))
			})
		}))

	Convey("An invalid call to download", t, withBinaryTestServer(http.StatusInternalServerError,
		"/v1/secure-file/test/file/hello.txt",
		http.MethodGet,
		"hello.txt",
		[]byte("hello world"),
		func(ts *httptest.Server) {
			cl, _ := NewClient(GenerateMockAuth(ts.URL, "a-cool-token", false, false), nil)
			So(cl, ShouldNotBeNil)
			Convey("Should return a valid file", func() {
				fileBuffer.Reset()
				err := cl.SecureFile().Get("/test/file/hello.txt", &fileBuffer)
				So(err, ShouldNotBeNil)
			})
		}))

	Convey("A download to a non-responsive server", t, func() {
		cl, _ := NewClient(GenerateMockAuth("http://127.0.0.1:32876", "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		Convey("Should return an error", func() {
			fileBuffer.Reset()
			err := cl.SecureFile().Get("/test/file/hello.txt", &fileBuffer)
			So(err, ShouldNotBeNil)
		})
	})
}

func getTestInputReader(t *testing.T, content string) io.Reader {
	var buf bytes.Buffer
	if _, err := buf.WriteString(content); err != nil {
		t.Fatalf("Error creating temp input: %v", err)
	}
	return &buf
}

func TestSecureFilePut(t *testing.T) {
	expectedContent := "hello world"

	Convey("A valid call to put", t, withBinaryTestServer(http.StatusNoContent,
		"/v1/secure-file/test/file/hello.txt",
		http.MethodPost,
		"hello.txt",
		[]byte(expectedContent),
		func(ts *httptest.Server) {
			cl, _ := NewClient(GenerateMockAuth(ts.URL, "a-cool-token", false, false), nil)
			So(cl, ShouldNotBeNil)
			Convey("Should return a valid file", func() {
				err := cl.SecureFile().Put(
					"/test/file/hello.txt",
					"hello.txt",
					getTestInputReader(t, expectedContent))
				So(err, ShouldBeNil)
			})
		}))

	Convey("An invalid call to put", t, withBinaryTestServer(http.StatusInternalServerError,
		"/v1/secure-file/test/file/hello.txt",
		http.MethodPost,
		"hello.txt",
		[]byte(expectedContent),
		func(ts *httptest.Server) {
			cl, _ := NewClient(GenerateMockAuth(ts.URL, "a-cool-token", false, false), nil)
			So(cl, ShouldNotBeNil)
			Convey("Should return a valid file", func() {
				err := cl.SecureFile().Put(
					"/test/file/hello.txt",
					"hello.txt",
					getTestInputReader(t, expectedContent))
				So(err, ShouldNotBeNil)
			})
		}))

	Convey("A put to a non-responsive server", t, func() {
		cl, _ := NewClient(GenerateMockAuth("http://127.0.0.1:32876", "a-cool-token", false, false), nil)
		So(cl, ShouldNotBeNil)
		Convey("Should return an error", func() {
			err := cl.SecureFile().Put(
				"/test/file/hello.txt",
				"hello.txt",
				getTestInputReader(t, expectedContent))
			So(err, ShouldNotBeNil)
		})
	})
}
