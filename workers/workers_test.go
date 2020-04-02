package workers

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/chapsuk/worker"
	"github.com/im-kulikov/helium/internal"
	"github.com/im-kulikov/helium/module"
	"github.com/im-kulikov/helium/service"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"go.uber.org/dig"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
)

type (
	config = map[string]interface{}

	testLocker   struct{ locked int32 }
	customLocker struct{ testLocker }

	hook struct{ called int }
)

const errFailToApplyLockerSettings = internal.Error("fail to apply locker settings")

func (h *hook) Caller(entry zapcore.Entry) error {
	if entry.Message == "run service" {
		h.called++
	}

	if entry.Message == "stop service" {
		h.called++
	}

	return nil
}

func mockedViper(cfg config) *viper.Viper {
	v := viper.New()

	v.SetDefault("workers.test", cfg)

	for key, val := range cfg {
		v.SetDefault("workers.test."+key, val)
	}

	return v
}

func (c *customLocker) Apply(key string, v *viper.Viper) (worker.Locker, error) {
	if v.GetBool(key + ".lock.fail") {
		return c, errFailToApplyLockerSettings
	}
	return c, nil
}

func (c *testLocker) Lock() error {
	if atomic.CompareAndSwapInt32(&c.locked, 0, 1) {
		return nil
	}
	return errors.New("locked")
}

func (c *testLocker) Unlock() {
	atomic.StoreInt32(&c.locked, 0)
}

func TestNewWorkers(t *testing.T) {
	tests := []struct {
		name string
		args Params
		err  error
		len  int
	}{
		{name: "empty config", err: ErrEmptyConfig},
		{
			name: "empty workers",
			err:  ErrEmptyWorkers,
			args: Params{Config: viper.New()},
		},
		{
			name: "missing worker key in config",
			err:  ErrMissingKey,
			args: Params{Config: viper.New(), Jobs: map[string]worker.Job{
				"missing_key_worker": func(ctx context.Context) {},
			}},
		},
		{
			name: "empty job",
			err:  ErrEmptyJob,
			args: Params{
				Jobs:   map[string]worker.Job{"test": nil},
				Config: mockedViper(nil),
			},
		},
		{
			name: "good case",
			len:  1,
			args: Params{
				Jobs:   map[string]worker.Job{"test": nil},
				Config: mockedViper(config{"disabled": true}),
			},
		},
	}
	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewWorkers(tt.args)
			if err != nil {
				require.Errorf(t, tt.err, err.Error())
				require.EqualError(t, errors.Cause(err), tt.err.Error())
				return
			}

			require.Len(t, got.Workers, tt.len)

			t.Run(tt.name+": check that all workers provided", func(t *testing.T) {
				di := dig.New()
				h := new(hook)

				// provide service.Module
				require.NoError(t, module.Provide(di, service.Module))
				// provide service workers
				require.NoError(t, di.Provide(func() Out { return got }))
				// provide testing logger
				require.NoError(t, di.Provide(func() *zap.Logger {
					return zaptest.NewLogger(t, zaptest.WrapOptions(
						zap.Hooks(h.Caller),
					))
				}))

				// try to receive service.Group
				require.NoError(t, di.Invoke(func(svc service.Group) {
					require.NoError(t, svc.Start(nil))
					svc.Stop()

					t.Logf("start/stop called %d times", h.called)

					require.Equal(t, len(got.Workers), h.called/2)
				}))
			})
		})
	}
}

func Test_workerByConfig(t *testing.T) {
	tests := []struct {
		name string
		args options
		err  error
	}{
		{name: "missing worker key", err: ErrMissingKey, args: options{
			Viper: mockedViper(config{}),
		}},

		{
			name: "missing locker",
			err:  ErrEmptyLocker,
			args: options{
				Name: "test",
				Job:  nopJob,
				Viper: mockedViper(config{
					"disabled":    false,
					"timer":       time.Millisecond,
					"ticker":      time.Millisecond,
					"cron":        "0 1 * * * *",
					"immediately": true,
					"lock":        true,
				}),
			},
		},

		{
			name: "error when apply locker settings",
			err:  errFailToApplyLockerSettings,
			args: options{
				Name:   "test",
				Job:    nopJob,
				Locker: &customLocker{},
				Viper: mockedViper(config{
					"disabled":    false,
					"timer":       time.Millisecond,
					"ticker":      time.Millisecond,
					"cron":        "0 1 * * * *",
					"immediately": true,
					"lock":        true,
					"lock.fail":   true,
				}),
			},
		},

		{
			name: "good case",
			args: options{
				Name: "test",
				Job:  nopJob,
				Viper: mockedViper(config{
					"disabled":    false,
					"timer":       time.Millisecond,
					"ticker":      time.Millisecond,
					"cron":        "0 1 * * * *",
					"immediately": true,
				}),
			},
		},

		{
			name: "good case with custom locker",
			args: options{
				Name:   "test",
				Job:    nopJob,
				Locker: &customLocker{},
				Viper: mockedViper(config{
					"disabled":    false,
					"timer":       time.Millisecond,
					"ticker":      time.Millisecond,
					"cron":        "0 1 * * * *",
					"immediately": true,
					"lock":        true,
				}),
			},
		},

		{
			name: "good case with simple locker",
			args: options{
				Name:   "test",
				Job:    nopJob,
				Locker: &testLocker{},
				Viper: mockedViper(config{
					"disabled":    false,
					"timer":       time.Millisecond,
					"ticker":      time.Millisecond,
					"cron":        "0 1 * * * *",
					"immediately": true,
					"lock":        true,
				}),
			},
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			got, err := workerByConfig(tt.args)
			if err != nil {
				require.Error(t, tt.err, err.Error())
				require.EqualError(t, errors.Cause(err), tt.err.Error())
				return
			}

			require.NotNil(t, got)
			require.NotPanics(t, func() { got.Run(ctx) })
		})
	}
}
