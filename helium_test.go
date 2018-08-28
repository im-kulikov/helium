package helium

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"testing"

	"github.com/bouk/monkey"
	"github.com/im-kulikov/helium/grace"
	"github.com/im-kulikov/helium/logger"
	"github.com/im-kulikov/helium/module"
	"github.com/im-kulikov/helium/settings"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/zap"
)

type (
	heliumApp    struct{}
	heliumErrApp struct{}
)

func (h heliumApp) Run(ctx context.Context) error    { return nil }
func (h heliumErrApp) Run(ctx context.Context) error { return errors.New("test") }

func TestHelium(t *testing.T) {
	Convey("Helium test suite", t, func() {
		Convey("create new helium without errors", func() {
			h, err := New(&settings.App{}, module.Module{
				{Constructor: func() App { return heliumApp{} }},
			}.Append(grace.Module, settings.Module, logger.Module))

			So(h, ShouldNotBeNil)
			So(err, ShouldBeNil)

			So(h.Run(), ShouldBeNil)
		})

		Convey("create new helium should fail on new", func() {
			h, err := New(&settings.App{}, module.Module{
				{Constructor: func() error { return nil }},
			})

			So(h, ShouldBeNil)
			So(err, ShouldBeError)
		})

		Convey("create new helium should fail on start", func() {
			h, err := New(&settings.App{}, module.Module{
				{Constructor: func() App { return heliumErrApp{} }},
			}.Append(grace.Module, settings.Module, logger.Module))

			So(h, ShouldNotBeNil)
			So(err, ShouldBeNil)

			So(h.Run(), ShouldBeError)
		})

		Convey("check catch", func() {
			var exitCode int
			monkey.Patch(os.Exit, func(code int) { exitCode = code })
			monkey.Patch(log.Fatal, func(...interface{}) { exitCode = 2 })
			defer monkey.Unpatch(os.Exit)

			Convey("should panic", func() {
				monkey.Patch(logger.NewLogger, func(*logger.Config, *settings.App) (*zap.Logger, error) {
					return nil, errors.New("test")
				})
				defer monkey.Unpatch(logger.NewLogger)

				err := errors.New("test")
				Catch(err)
				So(exitCode, ShouldEqual, 2)
			})

			Convey("should catch error", func() {
				monkey.Patch(fmt.Fprintf, func(io.Writer, string, ...interface{}) (int, error) {
					return 0, nil
				})
				defer monkey.Unpatch(fmt.Fprintf)

				err := errors.New("test")
				Catch(err)
				So(exitCode, ShouldEqual, 1)
			})

			Convey("shouldn't catch any", func() {
				Catch(nil)
				So(exitCode, ShouldBeZeroValue)
			})
		})
	})
}
