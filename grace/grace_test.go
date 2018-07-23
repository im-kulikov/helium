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

var _ = Convey

func TestGrace(t *testing.T) {
	Convey("check grace context", t, func() {
		var (
			log = zap.S()
			ctx = NewGracefulContext(log)
		)

		// waiting to run the goroutine and channel of signals
		<-time.Tick(100 * time.Millisecond)

		for _, sig := range []syscall.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP} {
			Convey(fmt.Sprintf("should cancel context on %s signal", sig), func() {
				err := syscall.Kill(syscall.Getpid(), sig)
				So(err, ShouldBeNil)

				select {
				case <-ctx.Done():
					So(true, ShouldBeTrue)
				case <-time.Tick(time.Second):
					So(errors.New("signal not catched"), ShouldBeNil)
				}
			})
		}
	})
}
