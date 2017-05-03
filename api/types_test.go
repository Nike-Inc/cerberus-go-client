package api

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestErrorResponse(t *testing.T) {
	var fakeError = ErrorResponse{
		ErrorID: "help-help-im-being-repressed",
		Errors: []ErrorDetail{
			ErrorDetail{
				Code:    12345,
				Message: "Strange women lying in ponds distributing swords is no basis for a system of government",
				Metadata: map[string]interface{}{
					"fields": "king_arthur",
				},
			},
		},
	}
	Convey("A valid ErrorResponse Error method call", t, func() {
		stringErr := fakeError.Error()
		Convey("Should give a valid string", func() {
			So(stringErr, ShouldStartWith, "Error from API. ID: help-help-im-being-repressed")
		})
	})
}
