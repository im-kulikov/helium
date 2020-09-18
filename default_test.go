package helium

import (
	"context"
	"testing"

	"github.com/im-kulikov/helium/logger"
	"github.com/im-kulikov/helium/module"
	"github.com/im-kulikov/helium/service"
	"github.com/im-kulikov/helium/settings"
	"github.com/stretchr/testify/require"
	"go.uber.org/dig"
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

func (e errService) Stop() error {
	if !e.stop {
		return nil
	}

	return testError
}

func (e errService) Name() string { return "errService" }

func TestDefaultApp(t *testing.T) {
	t.Run("create new helium with default application", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		h, err := New(&Settings{},
			DefaultApp,
			settings.Module,
			logger.Module,
			service.Module,
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
			settings.Module,
			logger.Module,
			service.Module,
			module.New(func() context.Context { return ctx }),
			module.New(func() service.Service { return errService{start: true} }, dig.Group("services")),
		)

		require.NotNil(t, h)
		require.NoError(t, err)

		cancel()

		require.EqualError(t, h.Run(), testError.Error())
	})

	t.Run("default application with stop err", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		h, err := New(&Settings{},
			DefaultApp,
			settings.Module,
			logger.Module,
			service.Module,
			module.New(func() context.Context { return ctx }),
			module.New(func() service.Service { return errService{stop: true} }, dig.Group("services")),
		)

		require.NotNil(t, h)
		require.NoError(t, err)

		cancel()

		require.EqualError(t, h.Run(), testError.Error())
	})
}
