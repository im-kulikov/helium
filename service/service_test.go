package service

import (
	"context"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
	"go.uber.org/dig"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"github.com/im-kulikov/helium/internal"
)

type (
	testWorker struct {
		*atomic.Error

		number  int
		errored bool
		started *atomic.Bool
	}

	testServiceOut struct {
		dig.Out
		Service Service `group:"services"`
	}

	testServicesOut struct {
		dig.Out

		// use `group:"servicesflatten"` to provide multiple
		// services
		Services []Service `group:"services,flatten"`
	}
)

const testError = internal.Error("test error")

var (
	iter = atomic.NewInt32(0)

	_ Service = (*testWorker)(nil)
)

func (t *testWorker) Start(ctx context.Context) error {
	if t.errored {
		return testError
	}

	t.started.Toggle()

	<-ctx.Done()

	return nil
}

func (t *testWorker) Stop(context.Context) {
	if t.errored {
		t.Store(testError)
	}

	t.started.Toggle()
}

func (t *testWorker) Name() string {
	return "test-worker-" + strconv.Itoa(t.number)
}

func newWorker() *testWorker {
	return &testWorker{
		number:  int(iter.Inc()),
		Error:   atomic.NewError(nil),
		started: atomic.NewBool(false),
	}
}

func TestServices(t *testing.T) {
	t.Run("should be ok", func(t *testing.T) {
		count := 10
		services := make([]Service, 0, count)

		for i := 0; i < count; i++ {
			services = append(services, newWorker())
		}

		// should ignore empty service
		services = append(services, nil)

		grp := newGroup(Params{
			Group:  services,
			Config: viper.New(),
			Logger: zaptest.NewLogger(t),
		})

		ctx, cancel := context.WithCancel(context.Background())

		group := new(sync.WaitGroup)
		start := make(chan struct{})

		group.Add(1)

		go func() {
			defer group.Done()

			<-start
			require.NoError(t, grp.Run(ctx))
		}()

		close(start)

		<-time.After(time.Millisecond * 5)

		for i := 0; i < count; i++ {
			if wrk, ok := services[i].(*testWorker); ok && !services[i].(*testWorker).started.Load() {
				t.Fatalf("worker(%d) should be started", wrk.number)
			}
		}

		cancel()
		group.Wait()
		for i := 0; i < count; i++ {
			require.False(t, services[i].(*testWorker).started.Load())
		}
	})

	t.Run("should panics on stop", func(t *testing.T) {
		wrk := newWorker()
		wrk.errored = true

		grp := newGroup(Params{
			Config: viper.New(),
			Group:  []Service{wrk},
			Logger: zaptest.NewLogger(t),
		})

		require.False(t, wrk.started.Load())

		// error should be passed from start
		require.EqualError(t, grp.Run(context.Background()), testError.Error())

		// error should be written on stop
		require.EqualError(t, wrk.Load(), testError.Error())
	})
}

func TestServicesFromDI(t *testing.T) {
	di := dig.New()
	cnt := 10

	// provide logger
	require.NoError(t, di.Provide(func() *zap.Logger {
		return zaptest.NewLogger(t)
	}))

	// provide service.Group
	require.NoError(t, di.Provide(newGroup))

	// provide single service by dig.Out
	require.NoError(t, di.Provide(func() testServiceOut {
		return testServiceOut{Service: newWorker()}
	}))

	// provide single service by return
	require.NoError(t, di.Provide(
		func() Service { return newWorker() },
		dig.Group("services"),
	))

	// provide multiple services by return
	require.NoError(t, di.Provide(
		func() []Service {
			return []Service{
				newWorker(),
				newWorker(),
			}
		},
		dig.Group("services,flatten"),
	))

	// provide multiple services
	require.NoError(t, di.Provide(func() testServicesOut {
		services := make([]Service, 0, cnt-1)
		for i := 0; i < cnt-4; i++ {
			services = append(services, newWorker())
		}
		return testServicesOut{Services: services}
	}))
}
