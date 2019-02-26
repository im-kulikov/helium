package logger

import (
	"os"
	"testing"

	"bou.ke/monkey"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/zap"
)

func TestStdLogger(t *testing.T) {
	Convey("StdLogger test suite", t, func(c C) {
		std := NewStdLogger(zap.L())

		c.Convey("Not fatal calls should not panic", func(c C) {
			c.Convey("print", func(c C) {
				c.So(func() { std.Print("panic no") }, ShouldNotPanic)
			})

			c.Convey("printf", func(c C) {
				c.So(func() { std.Printf("panic %s", "no") }, ShouldNotPanic)
			})

		})

		c.Convey("Fatal(f) should call os.Exit with 1 code", func(c C) {
			var exitCode int
			monkey.Patch(os.Exit, func(code int) { exitCode = code })
			defer monkey.Unpatch(os.Exit)

			c.Convey("Fatal", func(c C) {
				c.So(func() { std.Fatal("panic no") }, ShouldNotPanic)
				c.So(exitCode, ShouldEqual, 1)
			})

			c.Convey("Fatalf", func(c C) {
				c.So(func() { std.Fatalf("panic %s", "no") }, ShouldNotPanic)
				c.So(exitCode, ShouldEqual, 1)
			})
		})
	})
}
