package logger

import (
	"os"
	"testing"

	"github.com/bouk/monkey"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/zap"
)

func TestStdLogger(t *testing.T) {
	Convey("StdLogger test suite", t, func() {
		std := NewStdLogger(zap.L())

		Convey("Not fatal calls should not panic", func() {
			Convey("print", func() {
				So(func() { std.Print("panic no") }, ShouldNotPanic)
			})

			Convey("printf", func() {
				So(func() { std.Printf("panic %s", "no") }, ShouldNotPanic)
			})

		})

		Convey("Fatal(f) should call os.Exit with 1 code", func() {
			var exitCode int
			monkey.Patch(os.Exit, func(code int) { exitCode = code })
			defer monkey.Unpatch(os.Exit)

			Convey("Fatal", func() {
				So(func() { std.Fatal("panic no") }, ShouldNotPanic)
				So(exitCode, ShouldEqual, 1)
			})

			Convey("Fatalf", func() {
				So(func() { std.Fatalf("panic %s", "no") }, ShouldNotPanic)
				So(exitCode, ShouldEqual, 1)
			})
		})
	})
}
