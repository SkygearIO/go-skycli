package commands

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLocationValidate(t *testing.T) {
	Convey("Complex Location", t, func() {
		loc := newComplexLocation()

		Convey("gets an simpler form location", func() {
			input := "@loc:3,4"
			So(loc.Validate(input), ShouldBeTrue)
		})

		Convey("gets an random string", func() {
			input := "something"
			So(loc.Validate(input), ShouldBeFalse)
		})
	})
}

func TestLocationConvert(t *testing.T) {
	Convey("Complex Location", t, func() {
		loc := newComplexLocation()

		Convey("gets an simpler form location", func() {
			input := "@loc:3,4"
			expectedMap := map[string]interface{}{
				"$type": "geo",
				"$lat":  3.0,
				"$lng":  4.0,
			}

			output, err := loc.Convert(input)
			So(err, ShouldBeNil)
			outMap := output.(map[string]interface{})
			So(outMap, ShouldResemble, expectedMap)
		})

		Convey("gets an random string", func() {
			input := "something"

			_, err := loc.Convert(input)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestReferenceValidate(t *testing.T) {
	Convey("Complex Reference", t, func() {
		ref := newComplexReference()

		Convey("gets an simpler form reference", func() {
			input := "@ref:1234"
			So(ref.Validate(input), ShouldBeTrue)
		})

		Convey("gets an random string", func() {
			input := "something"
			So(ref.Validate(input), ShouldBeFalse)
		})
	})
}

func TestReferenceConvert(t *testing.T) {
	Convey("Complex Reference", t, func() {
		ref := newComplexReference()

		Convey("gets an simpler form reference", func() {
			input := "@ref:4321"
			expectedMap := map[string]interface{}{
				"$type": "ref",
				"$id":   "4321",
			}

			output, err := ref.Convert(input)
			So(err, ShouldBeNil)
			So(output, ShouldResemble, expectedMap)
		})

		Convey("gets an random string", func() {
			input := "something"

			_, err := ref.Convert(input)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestStringValidate(t *testing.T) {
	Convey("Complex String", t, func() {
		str := newComplexString()

		Convey("gets an simpler form string", func() {
			input := "@str:somestring"
			So(str.Validate(input), ShouldBeTrue)
		})

		Convey("gets an random string", func() {
			input := "something"
			So(str.Validate(input), ShouldBeFalse)
		})
	})
}

func TestStringConvert(t *testing.T) {
	Convey("Complex String", t, func() {
		str := newComplexString()

		Convey("gets an simpler form string", func() {
			input := "@str:somestring"
			expectedMap := map[string]interface{}{
				"$type": "str",
				"$str":  "somestring",
			}

			output, err := str.Convert(input)
			So(err, ShouldBeNil)
			So(output, ShouldResemble, expectedMap)
		})

		Convey("gets an random string", func() {
			input := "something"

			_, err := str.Convert(input)
			So(err, ShouldNotBeNil)
		})
	})
}
