package grace

import (
	"fmt"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestGrace(t *testing.T) {
	signals := []syscall.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP}
	for i := range signals {
		sig := signals[i]
		t.Run(fmt.Sprintf("should cancel context on %s signal", sig), func(t *testing.T) {
			is := assert.New(t)

			log := zaptest.NewLogger(t)
			ctx := NewGracefulContext(log)

			// waiting to run the goroutine and channel of signals
			<-time.After(time.Millisecond)

			err := syscall.Kill(syscall.Getpid(), sig)
			is.NoError(err)

			select {
			case <-ctx.Done():
				return
			case <-time.Tick(time.Second):
				t.Fatal("no signal")
			}
		})
	}
}
