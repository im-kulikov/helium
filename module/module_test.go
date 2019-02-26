package module

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/dig"
)

func TestModule(t *testing.T) {
	Convey("Test Module and Provider", t, func(c C) {
		c.Convey("Provider", func(c C) {
			var dic = dig.New()

			c.Convey("should return error on empty provider", func(c C) {
				p := new(Provider)
				err := Provide(dic, Module{p})
				c.So(err, ShouldBeError)
			})

			c.Convey("should return error if provider constructor func has to many similar returns values", func(c C) {
				mod := New(func() (int, int, error) {
					return 0, 0, nil
				})

				err := Provide(dic, mod)

				c.So(err, ShouldBeError)
			})

			c.Convey("should return error if provider constructor func has only error field", func(c C) {
				mod := New(func() error {
					return nil
				})

				err := Provide(dic, mod)

				c.So(err, ShouldBeError)
			})

			c.Convey("should not return errors on correct provider", func(c C) {
				mod := New(func() (int, error) {
					return 0, nil
				})

				err := Provide(dic, mod)
				c.So(err, ShouldBeNil)

				err = dic.Invoke(func(int) {})
				c.So(err, ShouldBeNil)
			})
		})

		c.Convey("Module", func(c C) {
			var (
				m1 = New(func() int32 { return 0 })
				m2 = New(func() int64 { return 1 })
				m3 = New(func() error { return nil })
				m4 = m1.Append(m2)
				m5 = m1.Append(m2, m3)

				dic = dig.New()
			)

			c.Convey("should create new module", func(c C) {
				c.So(m1, ShouldHaveLength, 1)
				c.So(m2, ShouldHaveLength, 1)
				c.So(m3, ShouldHaveLength, 1)
				c.So(m4, ShouldHaveLength, 2)
				c.So(m5, ShouldHaveLength, 3)
			})

			c.Convey("m1 and m2 should not fail", func(c C) {
				err := Provide(dic, m4)
				c.So(err, ShouldBeNil)

				err = dic.Invoke(func(int32, int64) {})
				c.So(err, ShouldBeNil)
			})

			c.Convey("m1 .. m3 should fail", func(c C) {
				err := Provide(dic, m5)
				c.So(err, ShouldBeError)
			})
		})
	})
}
