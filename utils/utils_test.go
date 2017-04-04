package utils

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
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
