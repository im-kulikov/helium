package nats

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/viper"
)

func TestNewDefaultConfig(t *testing.T) {
	Convey("Check nats module", t, func() {
		v := viper.New()

		Convey("must fail on empty", func() {
			c, err := NewDefaultConfig(v)
			So(c, ShouldBeNil)
			So(err, ShouldBeError)
		})

		Convey("should be ok", func() {
			url := "something"
			v.SetDefault("nats.url", url)

			c, err := NewDefaultConfig(v)
			So(err, ShouldBeNil)
			So(c.Url, ShouldEqual, url)
		})

		Convey("should fail for empty config", func() {
			c, err := NewConnection(nil)
			So(c, ShouldBeNil)
			So(err, ShouldBeError, ErrEmptyConfig)
		})
	})
}
