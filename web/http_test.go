package web

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc/test/bufconn"

	"github.com/im-kulikov/helium/group"
	"github.com/im-kulikov/helium/internal"
)

type logger struct {
	*zap.Logger
	*bytes.Buffer

	Result *testLogResult
}

type fakeWriteSyncer struct {
	io.Writer
}

var _ zapcore.WriteSyncer = (*fakeWriteSyncer)(nil)

func (f *fakeWriteSyncer) Sync() error { return nil }

type testLogResult struct {
	L, T, M string
	E       string `json:"error"`
	N       string `json:"name"`
}

func newTestLogger() *logger {
	buf := new(bytes.Buffer)

	return &logger{
		Buffer: buf,

		Logger: zap.New(zapcore.NewCore(
			zapcore.NewJSONEncoder(zap.NewDevelopmentEncoderConfig()),
			&fakeWriteSyncer{Writer: buf},
			zapcore.DebugLevel)),

		Result: new(testLogResult),
	}
}

func (tl *logger) Error() string {
	return tl.Result.E
}

func (tl *logger) Cleanup() {
	tl.Buffer.Reset()
	tl.Result = new(testLogResult)
}

func (tl *logger) Empty() bool {
	return tl.Buffer.String() == ""
}

func (tl *logger) Decode() error {
	return json.NewDecoder(tl.Buffer).Decode(&tl.Result)
}

func TestHTTPService(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	t.Run("should set network and address", func(t *testing.T) {
		lis, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		require.NoError(t, lis.Close())

		serve, err := NewHTTPService(
			&http.Server{},
			HTTPSkipErrors(),
			HTTPName(apiServer),
			HTTPListener(nil),
			HTTPWithLogger(nil),
			HTTPListenAddress(lis.Addr().String()),
			HTTPListenNetwork(lis.Addr().Network()))
		require.NoError(t, err)

		s, ok := serve.(*httpService)
		require.True(t, ok)
		require.Equal(t, lis.Addr().String(), s.address)
		require.Equal(t, lis.Addr().Network(), s.network)
	})

	t.Run("should fail on empty address", func(t *testing.T) {
		serve, err := NewHTTPService(&http.Server{})
		require.Nil(t, serve)
		require.EqualError(t, err, ErrEmptyHTTPAddress.Error())
	})

	t.Run("should fail on empty server", func(t *testing.T) {
		serve, err := NewHTTPService(nil)
		require.Nil(t, serve)
		require.EqualError(t, err, ErrEmptyHTTPServer.Error())
	})

	t.Run("should fail on Start and Stop", func(t *testing.T) {
		log := newTestLogger()

		require.EqualError(t, (&httpService{logger: log.Logger}).Start(ctx), ErrEmptyHTTPServer.Error())
		log.Cleanup()

		(&httpService{logger: log.Logger}).Stop(ctx)
		require.NoError(t, log.Decode())
		require.EqualError(t, log, ErrEmptyHTTPServer.Error())
	})

	t.Run("should fail on net.Listen", func(t *testing.T) {
		srv, err := NewHTTPService(&http.Server{}, HTTPListenAddress("test:80"))
		require.Nil(t, srv)
		require.Error(t, err)
		require.Contains(t, err.Error(), "listen tcp: lookup test")
	})

	t.Run("should fail for serve", func(t *testing.T) {
		lis := bufconn.Listen(listenSize)
		s := &http.Server{}

		serve, err := NewHTTPService(s,
			HTTPListener(lis),
			HTTPWithLogger(zaptest.NewLogger(t)))
		require.NoError(t, err)
		require.NotEmpty(t, serve.Name())

		top, stop := context.WithCancel(ctx)
		require.NoError(t, group.New().
			Add(serve.Start, func(context.Context) { serve.Stop(top) }).
			Add(func(ctx context.Context) error {
				stop()

				con, errConn := lis.Dial()
				if errConn != nil {
					return errConn
				}

				go func() {
					// emulate long query:
					time.Sleep(time.Second)

					_ = con.Close()
				}()

				return nil
			}, func(context.Context) {}).
			Run(top))
	})

	t.Run("should not fail for tls", func(t *testing.T) {
		lis := bufconn.Listen(listenSize)
		s := &http.Server{
			// nolint:gosec
			TLSConfig: &tls.Config{ // #nosec G402: TLS MinVersion too low.
				GetCertificate: func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
					return nil, internal.Error("test")
				},
			},
		}

		serve, err := NewHTTPService(s,
			HTTPListener(lis),
			HTTPWithLogger(zaptest.NewLogger(t)))
		require.NoError(t, err)
		require.NotEmpty(t, serve.Name())

		top, stop := context.WithCancel(ctx)
		require.NoError(t, group.New().
			Add(serve.Start, serve.Stop).
			Add(func(context.Context) error {
				stop()

				return nil
			}, func(context.Context) {}).
			Run(top))
	})
}
