package orm

import (
	"testing"

	"github.com/go-pg/pg"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func TestNewDefaultConfig(t *testing.T) {
	Convey("Check orm module", t, func() {
		v := viper.New()
		l := zap.L()

		Convey("must fail on empty", func() {
			c, err := NewDefaultConfig(v)
			So(c, ShouldBeNil)
			So(err, ShouldBeError)
		})

		Convey("should be ok", func() {
			url := "localhost"
			v.SetDefault("postgres.address", url)

			c, err := NewDefaultConfig(v)
			So(err, ShouldBeNil)
			So(c.Addr, ShouldEqual, url)
		})

		Convey("should fail for empty config", func() {
			c, err := NewConnection(nil, l)
			So(c, ShouldBeNil)
			So(err, ShouldBeError, ErrEmptyConfig)
		})

		Convey("should fail for empty logger", func() {
			url := "localhost"
			v.SetDefault("postgres.address", url)

			c, err := NewDefaultConfig(v)
			So(err, ShouldBeNil)
			So(c.Addr, ShouldEqual, url)

			cli, err := NewConnection(c, nil)
			So(cli, ShouldBeNil)
			So(err, ShouldBeError, ErrEmptyLogger)
		})

		Convey("should not connect", func() {
			v.SetDefault("postgres.username", "unknown")
			v.SetDefault("postgres.password", "postgres")
			v.SetDefault("postgres.database", "postgres")

			c, err := NewDefaultConfig(v)
			So(err, ShouldBeNil)

			cli, err := NewConnection(c, l)
			So(err, ShouldBeError)
			So(cli, ShouldBeNil)
		})

		Convey("should connect", func() {
			v.SetDefault("postgres.debug", true)
			v.SetDefault("postgres.username", "postgres")
			v.SetDefault("postgres.password", "postgres")
			v.SetDefault("postgres.database", "postgres")

			c, err := NewDefaultConfig(v)
			So(err, ShouldBeNil)
			So(c.User, ShouldEqual, "postgres")
			So(c.Password, ShouldEqual, "postgres")
			So(c.Database, ShouldEqual, "postgres")

			cli, err := NewConnection(c, l)
			So(err, ShouldBeNil)
			So(cli, ShouldNotBeNil)

			_, err = cli.ExecOne("SELECT 1")
			So(err, ShouldBeNil)

			err = cli.Close()
			So(err, ShouldBeNil)
		})

		Convey("should connect with before/after hooks", func() {
			v.SetDefault("postgres.debug", true)
			v.SetDefault("postgres.username", "postgres")
			v.SetDefault("postgres.password", "postgres")
			v.SetDefault("postgres.database", "postgres")

			c, err := NewDefaultConfig(v)
			So(err, ShouldBeNil)
			So(c.User, ShouldEqual, "postgres")
			So(c.Password, ShouldEqual, "postgres")
			So(c.Database, ShouldEqual, "postgres")

			cli, err := NewConnection(c, l)
			So(err, ShouldBeNil)
			So(cli, ShouldNotBeNil)

			cli.AddQueryHook(&Hook{
				Before: func(event *pg.QueryEvent) {},
				After:  nil,
			})

			_, err = cli.ExecOne("SELECT 1")
			So(err, ShouldBeNil)

			err = cli.Close()
			So(err, ShouldBeNil)
		})
	})
}
