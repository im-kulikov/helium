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

	actors []actor
	ignore []error
	period time.Duration

	expect error
}

const (
	errAlways = internal.Error("always")

	defaultAwait = time.Millisecond * 5
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

	{ // errored actor
		ctx, cancel := context.WithCancel(context.Background())
		cases = append(cases, testCase{
			name: "errored actor",

			ctx:    ctx,
			cancel: cancel,

			await: defaultAwait,

			period: time.Nanosecond,
			expect: errAlways,
			actors: []actor{
				{
					shutdown: func(ctx context.Context) { <-ctx.Done() },
					callback: func(context.Context) error { return errAlways },
				},
			},
		})
	}

	{ // should stop on context deadline
		ctx, cancel := context.WithTimeout(context.Background(), defaultAwait/3)
		cases = append(cases, testCase{
			name: "should stop on context deadline",

			ctx:    ctx,
			cancel: cancel,

			await: defaultAwait,

			period: defaultAwait / 3,
			actors: []actor{
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

			period: time.Nanosecond,

			actors: []actor{
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
				WithShutdownTimeout(tt.period),
				WithIgnoreErrors(tt.ignore...))

			for j := range tt.actors {
				run.Add(tt.actors[j].callback, tt.actors[j].shutdown)
			}

			require.Equal(t, tt.expect, run.Run(tt.ctx))
			require.LessOrEqual(t, time.Since(now), tt.await)
		})
	}
}
