package helium

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"bou.ke/monkey"
	"github.com/im-kulikov/helium/grace"
	"github.com/im-kulikov/helium/logger"
	"github.com/im-kulikov/helium/module"
	"github.com/im-kulikov/helium/settings"
	"github.com/stretchr/testify/require"
	"go.uber.org/dig"
	"go.uber.org/zap"
)

type (
	heliumApp    struct{}
	heliumErrApp struct{}

	Error string
)

const ErrTest = Error("test")

func (e Error) Error() string {
	return string(e)
}

func (h heliumApp) Run(ctx context.Context) error    { return nil }
func (h heliumErrApp) Run(ctx context.Context) error { return ErrTest }

func TestHelium(t *testing.T) {
	t.Run("create new helium without errors", func(t *testing.T) {
		h, err := New(&Settings{}, module.Module{
			{Constructor: func() App { return heliumApp{} }},
		}.Append(grace.Module, settings.Module, logger.Module))

		require.NotNil(t, h)
		require.NoError(t, err)
		require.NoError(t, h.Run())
	})

	t.Run("create new helium and setup ENV", func(t *testing.T) {
		tmpFile, err := ioutil.TempFile("", "example")
		require.NoError(t, err)

		defer func() {
			require.NoError(t, os.Remove(tmpFile.Name()))
		}() // clean up

		err = os.Setenv("ABC_CONFIG", tmpFile.Name())
		require.NoError(t, err)

		err = os.Setenv("ABC_CONFIG_TYPE", "toml")
		require.NoError(t, err)

		h, err := New(&Settings{
			Name: "Abc",
		}, module.Module{
			{Constructor: func(cfg *settings.Core) App {
				require.Equal(t, tmpFile.Name(), cfg.File)
				require.Equal(t, "toml", cfg.Type)
				return heliumApp{}
			}},
		}.Append(grace.Module, settings.Module, logger.Module))

		require.NotNil(t, h)
		require.NoError(t, err)
		require.NoError(t, h.Run())
	})

	t.Run("create new helium should fail on new", func(t *testing.T) {
		h, err := New(&Settings{}, module.Module{
			{Constructor: func() error { return nil }},
		})

		require.Nil(t, h)
		require.Error(t, err)
	})
	t.Run("create new helium should fail on start", func(t *testing.T) {
		h, err := New(&Settings{}, module.Module{
			{Constructor: func() App { return heliumErrApp{} }},
		}.Append(grace.Module, settings.Module, logger.Module))

		require.NotNil(t, h)
		require.NoError(t, err)

		require.Error(t, h.Run())
	})

	t.Run("create new helium should fail on start", func(t *testing.T) {
		h, err := New(&Settings{}, module.Module{
			{Constructor: func() App { return heliumErrApp{} }},
		}.Append(grace.Module, settings.Module, logger.Module))

		require.NotNil(t, h)
		require.NoError(t, err)

		require.Error(t, h.Run())
	})

	t.Run("invoke dependencies from helium container", func(t *testing.T) {
		t.Run("should be ok", func(t *testing.T) {
			h, err := New(&Settings{}, grace.Module.Append(settings.Module, logger.Module))

			require.NotNil(t, h)
			require.NoError(t, err)

			require.Nil(t, h.Invoke(func() {}))
		})

		t.Run("should fail", func(t *testing.T) {
			h, err := New(&Settings{}, grace.Module.Append(settings.Module, logger.Module))

			require.NotNil(t, h)
			require.NoError(t, err)

			require.Error(t, h.Invoke(func(string) {}))
		})
	})

	t.Run("check catch", func(t *testing.T) {
		t.Run("should panic", func(t *testing.T) {
			var exitCode int

			monkey.Patch(os.Exit, func(code int) { exitCode = code })
			monkey.Patch(log.Fatal, func(...interface{}) { exitCode = 2 })

			defer monkey.UnpatchAll()

			monkey.Patch(logger.NewLogger, func(*logger.Config, *settings.Core) (*zap.Logger, error) {
				return nil, errors.New("test")
			})
			defer monkey.Unpatch(logger.NewLogger)

			err := errors.New("test")
			Catch(err)
			require.Equal(t, 2, exitCode)
		})

		t.Run("should catch error", func(t *testing.T) {
			var exitCode int

			monkey.Patch(os.Exit, func(code int) { exitCode = code })
			monkey.Patch(log.Fatal, func(...interface{}) { exitCode = 2 })

			defer monkey.UnpatchAll()

			monkey.Patch(fmt.Fprintf, func(io.Writer, string, ...interface{}) (int, error) {
				return 0, nil
			})
			defer monkey.Unpatch(fmt.Fprintf)

			monkey.Patch(logger.NewLogger, func(*logger.Config, *settings.Core) (*zap.Logger, error) {
				return zap.NewNop(), nil
			})
			defer monkey.Unpatch(logger.NewLogger)

			err := errors.New("test")
			Catch(err)
			require.Equal(t, 1, exitCode)
		})

		t.Run("shouldn't catch any", func(t *testing.T) {
			var exitCode int

			monkey.Patch(os.Exit, func(code int) { exitCode = code })
			monkey.Patch(log.Fatal, func(...interface{}) { exitCode = 2 })

			defer monkey.UnpatchAll()

			Catch(nil)
			require.Empty(t, exitCode)
		})

		t.Run("should catch stacktrace simple error", func(t *testing.T) {
			var exitCode int

			monkey.Patch(os.Exit, func(code int) { exitCode = code })
			monkey.Patch(log.Fatal, func(...interface{}) { exitCode = 2 })

			defer monkey.UnpatchAll()

			monkey.Patch(fmt.Printf, func(string, ...interface{}) (int, error) {
				return 0, nil
			})

			require.Panics(t, func() {
				CatchTrace(
					errors.New("test"))
			})

			require.Empty(t, exitCode)
		})

		t.Run("should catch stacktrace on nil", func(t *testing.T) {
			var exitCode int

			monkey.Patch(os.Exit, func(code int) { exitCode = code })
			monkey.Patch(log.Fatal, func(...interface{}) { exitCode = 2 })

			defer monkey.UnpatchAll()

			require.NotPanics(t, func() {
				CatchTrace(nil)
			})

			require.Empty(t, exitCode)
		})

		t.Run("should catch stacktrace on dig.Errors", func(t *testing.T) {
			var exitCode int

			monkey.Patch(os.Exit, func(code int) { exitCode = code })
			monkey.Patch(log.Fatal, func(...interface{}) { exitCode = 2 })

			defer monkey.UnpatchAll()

			monkey.Patch(fmt.Fprintf, func(io.Writer, string, ...interface{}) (int, error) { return 0, nil })
			defer monkey.Unpatch(fmt.Fprintf)

			require.Panics(t, func() {
				di := dig.New()
				CatchTrace(di.Invoke(func(log *zap.Logger) error {
					return nil
				}))
			})

			require.Empty(t, exitCode)
		})

		t.Run("should catch context.DeadlineExceeded on dig.Errors", func(t *testing.T) {
			var exitCode int

			monkey.Patch(os.Exit, func(code int) { exitCode = code })
			monkey.Patch(log.Fatal, func(...interface{}) { exitCode = 2 })

			defer monkey.UnpatchAll()

			require.Panics(t, func() {
				di := dig.New()
				CatchTrace(di.Invoke(func() error {
					return context.DeadlineExceeded
				}))
			})

			require.Empty(t, exitCode)
		})
	})
}
