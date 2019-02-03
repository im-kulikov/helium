package web

import (
	"bytes"
	"os"
	"testing"

	"bou.ke/monkey"
	"github.com/labstack/gommon/log"
	. "github.com/smartystreets/goconvey/convey"
)

func TestLogger(t *testing.T) {
	Convey("Try logger", t, func() {
		var (
			b        = new(bytes.Buffer)
			z        = newTestLogger(testBuffer{Buffer: b})
			l        = NewLogger(z)
			exitCode = 0
		)

		monkey.Patch(os.Exit, func(code int) { exitCode = code })

		Convey("try nullWriter", func() {
			out := l.Output()
			size, err := out.Write(make([]byte, 10))
			So(out, ShouldEqual, Null)
			So(size, ShouldEqual, 10)
			So(err, ShouldBeNil)
		})

		Convey("try customize", func() {
			l.SetLevel(log.ERROR)  // do nothing
			l.SetOutput(os.Stdout) // do nothing
			l.SetPrefix("prefix")  // do nothing
			l.SetHeader("header")  // do nothing

			So(l.Prefix(), ShouldEqual, "")
			So(l.Level(), ShouldEqual, log.DEBUG)
		})

		Convey("try Print", func() {
			l.Print("")
			So(b.String(), ShouldContainSubstring, "info")
		})
		Convey("try Printf", func() {
			l.Printf("")
			So(b.String(), ShouldContainSubstring, "info")
		})
		Convey("try Printj", func() {
			l.Printj(log.JSON{})
			So(b.String(), ShouldContainSubstring, "info")
		})
		Convey("try Debug", func() {
			l.Debug("")
			So(b.String(), ShouldContainSubstring, "debug")
		})
		Convey("try Debugf", func() {
			l.Debugf("")
			So(b.String(), ShouldContainSubstring, "debug")
		})
		Convey("try Debugj", func() {
			l.Debugj(log.JSON{})
			So(b.String(), ShouldContainSubstring, "debug")
		})
		Convey("try Info", func() {
			l.Info("")
			So(b.String(), ShouldContainSubstring, "info")
		})
		Convey("try Infof", func() {
			l.Infof("")
			So(b.String(), ShouldContainSubstring, "info")
		})
		Convey("try Infoj", func() {
			l.Infoj(log.JSON{})
			So(b.String(), ShouldContainSubstring, "info")
		})
		Convey("try Warn", func() {
			l.Warn("")
			So(b.String(), ShouldContainSubstring, "warn")
		})
		Convey("try Warnf", func() {
			l.Warnf("")
			So(b.String(), ShouldContainSubstring, "warn")
		})
		Convey("try Warnj", func() {
			l.Warnj(log.JSON{})
			So(b.String(), ShouldContainSubstring, "warn")
		})
		Convey("try Error", func() {
			l.Error("")
			So(b.String(), ShouldContainSubstring, "error")
		})
		Convey("try Errorf", func() {
			l.Errorf("")
			So(b.String(), ShouldContainSubstring, "error")
		})
		Convey("try Errorj", func() {
			l.Errorj(log.JSON{})
			So(b.String(), ShouldContainSubstring, "error")
		})

		Convey("try Fatal", func() {
			l.Fatal("")
			So(exitCode, ShouldNotEqual, 2)
			So(b.String(), ShouldContainSubstring, "fatal")
		})
		Convey("try Fatalf", func() {
			l.Fatalf("")
			So(exitCode, ShouldNotEqual, 2)
			So(b.String(), ShouldContainSubstring, "fatal")
		})
		Convey("try Fatalj", func() {
			l.Fatalj(log.JSON{})
			So(exitCode, ShouldNotEqual, 2)
			So(b.String(), ShouldContainSubstring, "fatal")
		})

		Convey("try Panic", func() {
			So(func() {
				l.Panic("")
			}, ShouldPanic)
			So(exitCode, ShouldNotEqual, 2)
			So(b.String(), ShouldContainSubstring, "panic")
		})
		Convey("try Panicf", func() {
			So(func() {
				l.Panicf("")
			}, ShouldPanic)
			So(exitCode, ShouldNotEqual, 2)
			So(b.String(), ShouldContainSubstring, "panic")
		})
		Convey("try Panicj", func() {
			So(func() {
				l.Panicj(log.JSON{})
			}, ShouldPanic)
			So(exitCode, ShouldNotEqual, 2)
			So(b.String(), ShouldContainSubstring, "panic")
		})
	})
}
