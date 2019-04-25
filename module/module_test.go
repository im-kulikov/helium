package module

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/dig"
)

func TestModule(t *testing.T) {
	t.Run("Provider", func(t *testing.T) {
		t.Run("should return error on empty provider", func(t *testing.T) {
			var dic = dig.New()

			p := new(Provider)
			err := Provide(dic, Module{p})
			require.Error(t, err)
		})

		t.Run("should return error if provider constructor func has to many similar returns values", func(t *testing.T) {
			var dic = dig.New()

			mod := New(func() (int, int, error) {
				return 0, 0, nil
			})

			err := Provide(dic, mod)

			require.Error(t, err)
		})

		t.Run("should return error if provider constructor func has only error field", func(t *testing.T) {
			var dic = dig.New()

			mod := New(func() error {
				return nil
			})

			err := Provide(dic, mod)

			require.Error(t, err)
		})

		t.Run("should not return errors on correct provider", func(t *testing.T) {
			var dic = dig.New()

			mod := New(func() (int, error) {
				return 0, nil
			})

			err := Provide(dic, mod)
			require.NoError(t, err)

			err = dic.Invoke(func(int) {})
			require.NoError(t, err)
		})
	})

	t.Run("Module", func(t *testing.T) {
		var (
			m1 = New(func() int32 { return 0 })
			m2 = New(func() int64 { return 1 })
			m3 = New(func() error { return nil })
			m4 = m1.Append(m2)
			m5 = m1.Append(m2, m3)

			dic = dig.New()
		)

		t.Run("should create new module", func(t *testing.T) {
			require.Len(t, m1, 1)
			require.Len(t, m2, 1)
			require.Len(t, m3, 1)
			require.Len(t, m4, 2)
			require.Len(t, m5, 3)
		})

		t.Run("m1 and m2 should not fail", func(t *testing.T) {
			err := Provide(dic, m4)
			require.NoError(t, err)

			err = dic.Invoke(func(int32, int64) {})
			require.NoError(t, err)
		})

		t.Run("m1 .. m3 should fail", func(t *testing.T) {
			err := Provide(dic, m5)
			require.Error(t, err)
		})
	})
}
