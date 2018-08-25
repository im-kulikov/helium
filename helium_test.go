package helium

import (
	"context"
	"errors"
	"testing"

	"github.com/im-kulikov/helium/grace"
	"github.com/im-kulikov/helium/logger"
	"github.com/im-kulikov/helium/module"
	"github.com/im-kulikov/helium/settings"
	. "github.com/smartystreets/goconvey/convey"
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
	})
}
