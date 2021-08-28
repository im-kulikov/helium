package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestOpsDefaults(t *testing.T) {
	v := viper.New()

	OpsDefaults(v)

	require.Equal(t, v.GetString(cfgOpsAddress), opsDefaultAddress)
	require.Equal(t, v.GetString(cfgOpsNetwork), opsDefaultNetwork)

	require.False(t, v.GetBool(cfgOpsDisableMetrics))
	require.False(t, v.GetBool(cfgOpsDisableProfile))
	require.False(t, v.GetBool(cfgOpsDisableHealthy))

	keys := []string{
		cfgOpsReadTimeout,
		cfgOpsReadHeaderTimeout,
		cfgOpsWriteTimeout,
		cfgOpsIdleTimeout,
		cfgOpsMaxHeaderBytes,
	}

	for _, key := range keys {
		require.Empty(t, v.Get(key))
	}
}

func TestNewOpsConfig(t *testing.T) {
	var (
		cfg = viper.New()
		log = zap.NewNop()
	)

	cases := []struct {
		name string

		err error
		res *OpsConfig

		defaults func(v *viper.Viper)

		cfg *viper.Viper
		log *zap.Logger
	}{
		{name: "empty logger", cfg: viper.New(), err: ErrEmptyLogger},
		{name: "empty config", log: zap.NewNop(), err: ErrEmptyConfig},
		{name: "empty address", cfg: viper.New(), log: zap.NewNop(), err: ErrEmptyHTTPAddress},
		{name: "should be ok", cfg: cfg, log: log, defaults: OpsDefaults, res: &OpsConfig{
			HTTPConfig: HTTPConfig{
				Logger:  log,
				Name:    opsDefaultName,
				Address: opsDefaultAddress,
				Network: opsDefaultNetwork,
			},
		}},
	}

	for i := range cases {
		tt := cases[i]

		t.Run(tt.name, func(t *testing.T) {
			if tt.defaults != nil {
				tt.defaults(tt.cfg)
			}

			res, err := NewOpsConfig(tt.cfg, tt.log)
			switch {
			case tt.err != nil:
				require.EqualError(t, err, tt.err.Error())
			default:
				require.NoError(t, err)
				require.Equal(t, tt.res, res)
			}
		})
	}
}

func TestOpsConfig_probeChecker(t *testing.T) {
	probes := []ProbeChecker{
		func(context.Context) error { return ErrEmptyHTTPServer },
		func(context.Context) error { return nil },
		func(context.Context) error { return nil },
		nil,
	}

	cases := []struct {
		name string

		code int
		text string

		probes []ProbeChecker
	}{
		{name: "empty probes", code: http.StatusOK},
		{name: "should be ok", code: http.StatusOK, probes: probes[1:]},
		{
			name:   "should fail on first",
			probes: probes,
			text:   ErrEmptyHTTPServer.Error(),
			code:   http.StatusInternalServerError,
		},
	}

	for i := range cases {
		tt := cases[i]
		t.Run(tt.name, func(t *testing.T) {
			req := new(http.Request)
			rec := httptest.NewRecorder()

			probeChecker(tt.probes)(rec, req)

			require.Equal(t, tt.code, rec.Code)
			require.Equal(t, tt.text, strings.TrimSpace(rec.Body.String()))
		})
	}
}

func TestNewOpsServer(t *testing.T) {
	httpConfig := HTTPConfig{
		Logger:  zap.NewNop(),
		Name:    opsDefaultName,
		Address: "127.0.0.1:0",
		Network: opsDefaultNetwork,
	}

	queries := []string{
		opsPathMetrics,
		opsPathDebugVars,
		opsPathProfileIndex,
		opsPathProfileCMDLine,
		opsPathProfileProfile + "?seconds=1",
		opsPathProfileSymbol,
		opsPathProfileTrace + "?seconds=1",
		opsPathAppReady,
		opsPathAppHealthy,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	checker := func(status int) func(t *testing.T, handler http.Handler) {
		return func(t *testing.T, handler http.Handler) {
			t.Helper()

			for i := range queries {
				uri := queries[i]

				t.Run(uri, func(t *testing.T) {
					rec := httptest.NewRecorder()
					req := httptest.NewRequest(http.MethodGet, uri, nil)

					handler.ServeHTTP(rec, req)

					require.Equal(t, status, rec.Code)
				})
			}
		}
	}

	cases := []struct {
		name   string
		config OpsConfig

		checker func(*testing.T, http.Handler)
	}{
		{
			name: "disable all",
			config: OpsConfig{
				HTTPConfig:     httpConfig,
				DisableMetrics: true,
				DisableProfile: true,
				DisableHealthy: true,
			},
			checker: checker(http.StatusNotFound),
		},

		{
			name:    "enable all",
			config:  OpsConfig{HTTPConfig: httpConfig},
			checker: checker(http.StatusOK),
		},
	}

	for i := range cases {
		tt := cases[i]

		t.Run(tt.name, func(t *testing.T) {
			svc, err := NewOpsServer(&tt.config, OpsProbeParams{})
			require.NoError(t, err)
			require.IsType(t, &httpService{}, svc)

			defer svc.Stop(ctx)

			handler := svc.(*httpService).server.Handler
			require.NotEmpty(t, handler)

			if tt.checker == nil {
				return
			}

			tt.checker(t, handler)
		})
	}
}
