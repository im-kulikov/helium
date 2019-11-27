package web

import (
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"testing"
)

type fakeService struct {
	startError error
	stopError  error
}

func (f fakeService) Start() error {
	return f.startError
}

func (f fakeService) Stop() error {
	return f.stopError
}

func TestMultiService(t *testing.T) {
	t.Run("fail on empty logger", func(t *testing.T) {
		svc, err := New(nil)
		require.Nil(t, svc)
		require.EqualError(t, err, ErrEmptyLogger.Error())
	})

	t.Run("fail on empty services", func(t *testing.T) {
		svc, err := New(zap.L(), nil, nil)
		require.Nil(t, svc)
		require.EqualError(t, err, ErrEmptyServices.Error())
	})

	t.Run("should fail on start and return first error", func(t *testing.T) {
		svc, err := New(zap.L(),
			&fakeService{startError: ErrEmptyServices},
			&fakeService{startError: ErrEmptyLogger})
		require.NoError(t, err)
		require.EqualError(t, svc.Start(), ErrEmptyServices.Error())
	})

	t.Run("should fail on stop and return last error", func(t *testing.T) {
		l, err := zap.NewDevelopment()
		require.NoError(t, err)

		svc, err := New(l,
			&fakeService{stopError: ErrEmptyServices},
			&fakeService{stopError: ErrEmptyLogger})
		require.NoError(t, err)
		require.EqualError(t, svc.Stop(), ErrEmptyLogger.Error())
	})
}
