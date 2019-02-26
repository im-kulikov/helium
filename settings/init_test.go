package settings

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/viper"
)

func TestInit(t *testing.T) {
	Convey("Core settings test suite", t, func(c C) {
		cfg := &Core{}

		c.Convey("should be ok without file", func(c C) {
			v, err := New(cfg)
			c.So(err, ShouldBeNil)
			c.So(v, ShouldHaveSameTypeAs, viper.New())
		})

		c.Convey("should be ok with temp file", func(c C) {
			tmpFile, err := ioutil.TempFile("", "example")
			if err != nil {
				log.Fatal(err)
			}

			defer os.Remove(tmpFile.Name()) // clean up

			cfg.File = tmpFile.Name()
			v, err := New(cfg)
			c.So(err, ShouldBeNil)
			c.So(v, ShouldHaveSameTypeAs, viper.New())
		})

		c.Convey("should fail", func(c C) {
			cfg.File = "unknown file"
			v, err := New(cfg)
			c.So(err, ShouldBeError)
			c.So(v, ShouldBeNil)
		})
	})
}
