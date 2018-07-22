package module

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/dig"
)

var _ = Convey

func TestModule(t *testing.T) {
	Convey("Test Module and Provider", t, func() {
		Convey("Provider", func() {
			var dic = dig.New()

			Convey("should return error on empty provider", func() {
				p := new(Provider)
				err := Provide(dic, Module{p})
				So(err, ShouldBeError)
			})

			Convey("should return error if provider constructor func has to many similar returns values", func() {
				p := &Provider{Constructor: func() (int, int, error) {
					return 0, 0, nil
				}}

				err := Provide(dic, Module{p})

				So(err, ShouldBeError)
			})

			Convey("should return error if provider constructor func has only error field", func() {
				p := &Provider{Constructor: func() error {
					return nil
				}}

				err := Provide(dic, Module{p})

				So(err, ShouldBeError)
			})

			Convey("should not return errors on correct provider", func() {
				p := &Provider{Constructor: func() (int, error) {
					return 0, nil
				}}

				err := Provide(dic, Module{p})
				So(err, ShouldBeNil)

				err = dic.Invoke(func(int) {})
				So(err, ShouldBeNil)
			})
		})

		Convey("Module", func() {
			var (
				m1 = Module{
					{Constructor: func() int32 { return 0 }},
				}
				m2 = Module{
					{Constructor: func() int64 { return 1 }},
				}
				m3 = Module{
					{Constructor: func() error { return nil }},
				}

				dic = dig.New()
			)

			Convey("should create new module", func() {
				m4 := m1.Append(m2).Append(m3)

				So(m1, ShouldHaveLength, 1)
				So(m2, ShouldHaveLength, 1)
				So(m3, ShouldHaveLength, 1)
				So(m4, ShouldHaveLength, 3)
			})

			Convey("m1 and m2 should not fail", func() {
				err := Provide(dic, m1.Append(m2))
				So(err, ShouldBeNil)

				err = dic.Invoke(func(int32, int64) {})
				So(err, ShouldBeNil)
			})

			Convey("m1 .. m3 should fail", func() {
				err := Provide(dic, m1.Append(m2).Append(m3))
				So(err, ShouldBeError)
			})
		})
	})
}
