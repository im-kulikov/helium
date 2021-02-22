// +build !go1.16

package grace

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

// NewGracefulContext returns graceful context
func NewGracefulContext(l *zap.Logger) context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
		sig := <-ch
		l.Info("received stop signal",
			zap.String("signal", sig.String()))
		cancel()
	}()
	return ctx
}
