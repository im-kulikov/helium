package grace

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/im-kulikov/helium/module"
	"go.uber.org/zap"
)

// Module graceful context
var Module = module.Module{
	{Constructor: NewGracefulContext},
}

// NewGracefulContext returns graceful context
func NewGracefulContext(l *zap.SugaredLogger) context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
		sig := <-ch
		l.Infof("received signal: %s", sig.String())
		cancel()
	}()
	return ctx
}
