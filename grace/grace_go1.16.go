//go:build go1.16
// +build go1.16

package grace

import (
	"context"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

// NewGracefulContext returns graceful context.
func NewGracefulContext(l *zap.Logger) context.Context {
	ctx, _ := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	go func() {
		<-ctx.Done()
		l.Info("receive stop signal")
	}()

	return ctx
}
