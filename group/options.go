package group

import (
	"time"
)

// Option allows to change group settings.
type Option func(*group)

// WithShutdownTimeout allows to change shutdown period.
func WithShutdownTimeout(v time.Duration) Option {
	return func(g *group) {
		if v == 0 {
			return
		}

		g.period = v
	}
}

// WithIgnoreErrors allows to add ignored errors.
func WithIgnoreErrors(v ...error) Option {
	return func(g *group) { g.ignore = append(g.ignore, v...) }
}
