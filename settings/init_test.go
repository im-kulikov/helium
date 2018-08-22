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
	Convey("App settings test suite", t, func() {
		cfg := &App{}

		Convey("should be ok without file", func() {
			v, err := New(cfg)
			So(err, ShouldBeNil)
			So(v, ShouldHaveSameTypeAs, viper.New())
		})

		Convey("should be ok with temp file", func() {
			tmpFile, err := ioutil.TempFile("", "example")
			if err != nil {
				log.Fatal(err)
			}

			defer os.Remove(tmpFile.Name()) // clean up

			cfg.File = tmpFile.Name()
			v, err := New(cfg)
			So(err, ShouldBeNil)
			So(v, ShouldHaveSameTypeAs, viper.New())
		})

		Convey("should fail", func() {
			cfg.File = "unknown file"
			v, err := New(cfg)
			So(err, ShouldBeError)
			So(v, ShouldBeNil)
		})
	})
}
