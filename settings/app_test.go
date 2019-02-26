package settings

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type providerType = func() *Core

func TestApp(t *testing.T) {
	Convey("Settings test suite", t, func(c C) {
		cfg := &Core{}

		c.Convey("check provider", func(c C) {
			provider := cfg.Provider()
			c.So(provider, ShouldNotBeNil)
			c.So(provider.Constructor, ShouldHaveSameTypeAs, providerType(nil))
			appProvider := provider.Constructor.(providerType)
			c.So(appProvider(), ShouldEqual, cfg)
		})

		c.Convey("safe type", func(c C) {
			cases := []string{"bad", "toml", "yml", "yaml"}
			for _, item := range cases {
				cfg.Type = item

				if item == "bad" {
					item = "yml"
				}

				c.So(cfg.SafeType(), ShouldEqual, item)
			}
		})
	})
}
