package grace

import (
	"fmt"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestGrace(t *testing.T) {
	var (
		log = zap.L()
		ctx = NewGracefulContext(log)
	)

	// waiting to run the goroutine and channel of signals
	<-time.Tick(100 * time.Millisecond)

	for _, sig := range []syscall.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP} {
		t.Run(fmt.Sprintf("should cancel context on %s signal", sig), func(t *testing.T) {
			is := assert.New(t)

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
