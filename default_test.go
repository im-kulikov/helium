package helium

import (
	"context"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"go.uber.org/dig"
	"go.uber.org/zap"

	"github.com/im-kulikov/helium/module"
	"github.com/im-kulikov/helium/service"
)

type errService struct {
	start bool
	stop  bool
}

func (e errService) Start(_ context.Context) error {
	if !e.start {
		return nil
	}

	return testError
}

func (e errService) Stop(context.Context) {
	if e.stop {
		panic(testError)
	}
}

func (e errService) Name() string { return "errService" }

func TestDefaultApp(t *testing.T) {
	t.Run("create new helium with default application", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		h, err := New(&Settings{},
			DefaultApp,
			module.New(viper.New),
			module.New(zap.NewNop),
			module.New(func() context.Context { return ctx }),
		)

		require.NotNil(t, h)
		require.NoError(t, err)

		cancel()

		require.NoError(t, h.Run())
	})

	t.Run("default application with start err", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		h, err := New(&Settings{},
			DefaultApp,
			module.New(viper.New),
			module.New(zap.NewNop),
			module.New(func() context.Context { return ctx }),
			module.New(func() service.Service { return errService{start: true} }, dig.Group("services")),
		)

		require.NotNil(t, h)
		require.NoError(t, err)

		require.EqualError(t, h.Run(), testError.Error())

		cancel()
	})

	t.Run("default application with stop err", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		h, err := New(&Settings{},
			DefaultApp,
			module.New(viper.New),
			module.New(zap.NewNop),
			module.New(func() context.Context { return ctx }),
			module.New(func() service.Service { return errService{stop: true} }, dig.Group("services")),
		)

		require.NotNil(t, h)
		require.NoError(t, err)

		cancel()
		require.Panics(t, func() {
			t.Helper()

			require.NoError(t, h.Run())
		}, testError.Error())
	})
}
