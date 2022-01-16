package group

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/im-kulikov/helium/internal"
)

type testCase struct {
	name string

	ctx    context.Context
	cancel context.CancelFunc

	await time.Duration

	ignore   []error
	services []service
	shutdown time.Duration

	expect error
}

const (
	errAlways = internal.Error("always")

	defaultAwait = time.Millisecond * 10
)

func TestNew(t *testing.T) {
	var cases []testCase

	{ // empty
		ctx, cancel := context.WithCancel(context.Background())
		cases = append(cases, testCase{
			name: "empty",

			ctx:    ctx,
			cancel: cancel,

			await: defaultAwait,
		})
	}

	{ // should stop all services when one is failed
		ctx, cancel := context.WithCancel(context.Background())
		cases = append(cases, testCase{
			name: "should stop all services when one is failed",

			ctx:    ctx,
			cancel: cancel,

			await: defaultAwait,

			shutdown: time.Nanosecond,
			expect:   errAlways,
			services: []service{
				{
					shutdown: func(ctx context.Context) { <-ctx.Done() },
					callback: func(context.Context) error { return errAlways },
				},

				{
					shutdown: func(ctx context.Context) { <-ctx.Done() },
					callback: func(ctx context.Context) error {
						// should not freeze
						<-ctx.Done()

						return ctx.Err()
					},
				},
			},
		})
	}

	{ // errored service
		ctx, cancel := context.WithCancel(context.Background())
		cases = append(cases, testCase{
			name: "errored service",

			ctx:    ctx,
			cancel: cancel,

			await: defaultAwait,

			shutdown: time.Nanosecond,
			expect:   errAlways,
			services: []service{
				{
					shutdown: func(ctx context.Context) { <-ctx.Done() },
					callback: func(context.Context) error { return errAlways },
				},
			},
		})
	}

	{ // should stop on context deadline
		ctx, cancel := context.WithTimeout(context.Background(), defaultAwait/4)
		cases = append(cases, testCase{
			name: "should stop on context deadline",

			ctx:    ctx,
			cancel: cancel,

			await: defaultAwait,

			shutdown: defaultAwait / 4,
			services: []service{
				{
					shutdown: func(ctx context.Context) { <-ctx.Done() },
					callback: func(ctx context.Context) error {
						<-ctx.Done()
						return ctx.Err()
					},
				},
				{
					shutdown: func(ctx context.Context) { <-ctx.Done() },
					callback: func(ctx context.Context) error {
						<-ctx.Done()
						return ctx.Err()
					},
				},
			},
		})
	}

	{ // should stop on context canceled
		ctx, cancel := context.WithCancel(context.Background())

		cases = append(cases, testCase{
			name: "should stop on context",

			ctx:    ctx,
			cancel: cancel,

			await: defaultAwait,

			shutdown: time.Nanosecond,

			services: []service{
				{
					shutdown: func(ctx context.Context) { <-ctx.Done() },
					callback: func(ctx context.Context) error {
						cancel()
						return ctx.Err()
					},
				},
				{
					shutdown: func(ctx context.Context) { <-ctx.Done() },
					callback: func(ctx context.Context) error {
						<-ctx.Done()
						return ctx.Err()
					},
				},
			},
		})
	}

	for i := range cases {
		tt := cases[i]

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			now := time.Now()

			run := New(
				WithShutdownTimeout(tt.shutdown),
				WithIgnoreErrors(tt.ignore...))

			for j := range tt.services {
				run.Add(tt.services[j].callback, tt.services[j].shutdown)
			}

			require.Equal(t, tt.expect, run.Run(tt.ctx))
			require.LessOrEqual(t, time.Since(now), tt.await)
		})
	}
}
