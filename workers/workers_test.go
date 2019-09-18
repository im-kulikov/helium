package workers

import (
	"context"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"github.com/chapsuk/worker"
	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

type (
	config = map[string]interface{}

	testLocker   struct{ locked int32 }
	customLocker struct{ testLocker }
)

const errFailToApplyLockerSettings = Error("fail to apply locker settings")

func mockedViper(cfg config) *viper.Viper {
	v := viper.New()

	v.SetDefault("workers.test", cfg)

	for key, val := range cfg {
		v.SetDefault("workers.test."+key, val)
	}
	spew.Dump(v.AllSettings())
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

			require.Len(t, got, tt.len)
			require.NotPanics(t, func() {
				for i := range got {
					got[i].Run(context.Background())
				}
			})
		})
	}
}

func TestNewWorkersGroup(t *testing.T) {
	testWorkers := []*worker.Worker{worker.New(nil)}

	tests := []struct {
		name string
		want *worker.Group
		args []*worker.Worker
	}{
		{name: "empty workers", want: NewWorkersGroup(nil)},
		{
			name: "single worker",
			args: testWorkers,
			want: NewWorkersGroup(testWorkers),
		},
	}
	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			if got := NewWorkersGroup(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewWorkersGroup() = %v, want %v", got, tt.want)
			}
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
