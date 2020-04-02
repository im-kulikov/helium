package service

import (
	"context"
	"strconv"
	"testing"

	"github.com/im-kulikov/helium/internal"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
	"go.uber.org/dig"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

type (
	testWorker struct {
		number  int
		errored bool
		started bool
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

func (t *testWorker) Start(context.Context) error {
	if t.errored {
		return testError
	}
	t.started = true
	return nil
}

func (t *testWorker) Stop() {
	t.started = false
}

func (t *testWorker) Name() string {
	return "test-worker-" + strconv.Itoa(t.number)
}

func newWorker() *testWorker {
	return &testWorker{
		number:  int(iter.Inc()),
		started: false,
	}
}

func TestServices(t *testing.T) {
	count := 10
	services := make([]Service, 0, count)

	for i := 0; i < count; i++ {
		services = append(services, newWorker())
	}

	params := Params{
		Logger: zaptest.NewLogger(t),
		Group:  services,
	}

	{ // good case
		svc := newGroup(params)

		require.NoError(t, svc.Start(nil))
		for i := 0; i < count; i++ {
			require.True(t, services[i].(*testWorker).started)
		}

		svc.Stop()
		for i := 0; i < count; i++ {
			require.False(t, services[i].(*testWorker).started)
		}
	}

	{ // bad case
		wrk := newWorker()
		wrk.errored = true

		svc := newGroup(Params{
			Logger: params.Logger,
			Group:  []Service{wrk},
		})

		require.EqualError(t, svc.Start(nil), testError.Error())
		for i := 0; i < count; i++ {
			require.False(t, services[i].(*testWorker).started)
		}
	}
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

	// should provide 10 services
	require.NoError(t, di.Invoke(func(svc Group) {
		require.Len(t, svc.(*services).items, cnt)
	}))
}
