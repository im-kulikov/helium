package helium

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"bou.ke/monkey"
	"github.com/im-kulikov/helium/grace"
	"github.com/im-kulikov/helium/logger"
	"github.com/im-kulikov/helium/module"
	"github.com/im-kulikov/helium/settings"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/dig"
	"go.uber.org/zap"
)

type (
	heliumApp    struct{}
	heliumErrApp struct{}
)

func (h heliumApp) Run(ctx context.Context) error    { return nil }
func (h heliumErrApp) Run(ctx context.Context) error { return errors.New("test") }

func TestHelium(t *testing.T) {
	Convey("Helium test suite", t, func(c C) {
		c.Convey("create new helium without errors", func(c C) {
			h, err := New(&Settings{}, module.Module{
				{Constructor: func() App { return heliumApp{} }},
			}.Append(grace.Module, settings.Module, logger.Module))

			c.So(h, ShouldNotBeNil)
			c.So(err, ShouldBeNil)

			c.So(h.Run(), ShouldBeNil)
		})

		c.Convey("create new helium and setup ENV", func(c C) {
			tmpFile, err := ioutil.TempFile("", "example")
			if err != nil {
				log.Fatal(err)
			}

			defer os.Remove(tmpFile.Name()) // clean up

			os.Setenv("HELIUM_CONFIG", tmpFile.Name())
			os.Setenv("HELIUM_CONFIG_TYPE", "toml")

			h, err := New(&Settings{}, module.Module{
				{Constructor: func(cfg *settings.Core) App {
					c.So(cfg.File, ShouldEqual, tmpFile.Name())
					c.So(cfg.Type, ShouldEqual, "toml")
					return heliumApp{}
				}},
			}.Append(grace.Module, settings.Module, logger.Module))

			c.So(h, ShouldNotBeNil)
			c.So(err, ShouldBeNil)

			c.So(h.Run(), ShouldBeNil)
		})

		c.Convey("create new helium should fail on new", func(c C) {
			h, err := New(&Settings{}, module.Module{
				{Constructor: func() error { return nil }},
			})

			c.So(h, ShouldBeNil)
			c.So(err, ShouldBeError)
		})

		c.Convey("create new helium should fail on start", func(c C) {
			h, err := New(&Settings{}, module.Module{
				{Constructor: func() App { return heliumErrApp{} }},
			}.Append(grace.Module, settings.Module, logger.Module))

			c.So(h, ShouldNotBeNil)
			c.So(err, ShouldBeNil)

			c.So(h.Run(), ShouldBeError)
		})

		c.Convey("invoke dependencies from helium container", func(c C) {
			c.Convey("should be ok", func(c C) {
				h, err := New(&Settings{}, grace.Module.Append(settings.Module, logger.Module))

				c.So(h, ShouldNotBeNil)
				c.So(err, ShouldBeNil)

				c.So(h.Invoke(func() {}), ShouldBeNil)
			})

			c.Convey("should fail", func(c C) {
				h, err := New(&Settings{}, grace.Module.Append(settings.Module, logger.Module))

				c.So(h, ShouldNotBeNil)
				c.So(err, ShouldBeNil)

				c.So(h.Invoke(func(string) {}), ShouldBeError)
			})
		})

		c.Convey("check catch", func(c C) {
			var exitCode int

			monkey.Patch(os.Exit, func(code int) { exitCode = code })
			monkey.Patch(log.Fatal, func(...interface{}) { exitCode = 2 })

			defer monkey.UnpatchAll()

			c.Convey("should panic", func(c C) {
				monkey.Patch(logger.NewLogger, func(*logger.Config, *settings.Core) (*zap.Logger, error) {
					return nil, errors.New("test")
				})
				defer monkey.Unpatch(logger.NewLogger)

				err := errors.New("test")
				Catch(err)
				c.So(exitCode, ShouldEqual, 2)
			})

			c.Convey("should catch error", func(c C) {
				monkey.Patch(fmt.Fprintf, func(io.Writer, string, ...interface{}) (int, error) {
					return 0, nil
				})
				defer monkey.Unpatch(fmt.Fprintf)

				err := errors.New("test")
				Catch(err)
				c.So(exitCode, ShouldEqual, 1)
			})

			c.Convey("shouldn't catch any", func(c C) {
				Catch(nil)
				c.So(exitCode, ShouldBeZeroValue)
			})

			c.Convey("should catch stacktrace simple error", func(c C) {
				monkey.Patch(fmt.Printf, func(string, ...interface{}) (int, error) {
					return 0, nil
				})

				c.So(func() {
					CatchTrace(
						errors.New("test"))
				}, ShouldPanic)

				c.So(exitCode, ShouldEqual, 0)
			})

			c.Convey("should catch stacktrace on nil", func(c C) {
				c.So(func() {
					CatchTrace(nil)
				}, ShouldNotPanic)

				c.So(exitCode, ShouldEqual, 0)
			})

			c.Convey("should catch stacktrace on dig.Errors", func(c C) {
				c.So(func() {
					di := dig.New()
					CatchTrace(di.Invoke(func(log *zap.Logger) error {
						return nil
					}))
				}, ShouldPanic)

				c.So(exitCode, ShouldEqual, 0)
			})
		})
	})
}
