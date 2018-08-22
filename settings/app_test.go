package settings

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type providerType = func() *App

func TestApp(t *testing.T) {
	Convey("Settings test suite", t, func() {
		cfg := &App{}

		Convey("check provider", func() {
			provider := cfg.Provider()
			So(provider, ShouldNotBeNil)
			So(provider.Constructor, ShouldHaveSameTypeAs, providerType(nil))
			appProvider := provider.Constructor.(providerType)
			So(appProvider(), ShouldEqual, cfg)
		})

		Convey("safe type", func() {
			cases := []string{"bad", "toml", "yml", "yaml"}
			for _, item := range cases {
				cfg.Type = item

				if item == "bad" {
					item = "yml"
				}

				So(cfg.SafeType(), ShouldEqual, item)
			}
		})
	})
}
