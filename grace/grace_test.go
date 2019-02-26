package grace

import (
	"errors"
	"fmt"
	"syscall"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/zap"
)

func TestGrace(t *testing.T) {
	Convey("check grace context", t, func(c C) {
		var (
			log = zap.L()
			ctx = NewGracefulContext(log)
		)

		// waiting to run the goroutine and channel of signals
		<-time.Tick(100 * time.Millisecond)

		for _, sig := range []syscall.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP} {
			c.Convey(fmt.Sprintf("should cancel context on %s signal", sig), func(c C) {
				err := syscall.Kill(syscall.Getpid(), sig)
				c.So(err, ShouldBeNil)

				select {
				case <-ctx.Done():
					c.So(true, ShouldBeTrue)
				case <-time.Tick(time.Second):
					c.So(errors.New("no signal"), ShouldBeNil)
				}
			})
		}
	})
}
