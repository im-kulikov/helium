package web

import (
	"context"
	"expvar"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"go.uber.org/dig"
	"go.uber.org/zap"

	"github.com/im-kulikov/helium/internal"
	"github.com/im-kulikov/helium/module"
	"github.com/im-kulikov/helium/service"
)

// ProbeChecker used by ops-server ready and health handler.
type ProbeChecker func(context.Context) error

// HTTPConfig .
type HTTPConfig struct {
	Logger  *zap.Logger  `mapstructure:"-"`
	Handler http.Handler `mapstructure:"-"`

	Name    string `mapstructure:"name"`
	Address string `mapstructure:"address"`
	Network string `mapstructure:"network"`

	ReadTimeout       time.Duration `mapstructure:"read_timeout"`
	ReadHeaderTimeout time.Duration `mapstructure:"read_header_timeout"`
	WriteTimeout      time.Duration `mapstructure:"write_timeout"`
	IdleTimeout       time.Duration `mapstructure:"idle_timeout"`
	MaxHeaderBytes    int           `mapstructure:"max_header_bytes"`
}

// OpsConfig .
type OpsConfig struct {
	HTTPConfig `mapstructure:",squash"`

	DisableMetrics bool `mapstructure:"disable_metrics"`
	DisableProfile bool `mapstructure:"disable_pprof"`
	DisableHealthy bool `mapstructure:"disable_healthy"`
}

// OpsProbeParams allows setting health and ready probes for ops server.
type OpsProbeParams struct {
	dig.In

	HealthProbes []ProbeChecker `group:"health_probes"`
	ReadyProbes  []ProbeChecker `group:"ready_probes"`
}

const (
	// ErrEmptyConfig is raised when empty configuration passed into functions that requires it.
	ErrEmptyConfig = internal.Error("empty configuration")

	opsDefaultName = "ops-server"

	opsDefaultAddress = ":8081"
	opsDefaultNetwork = "tcp"

	cfgOpsAddress           = "ops.address"
	cfgOpsNetwork           = "ops.network"
	cfgOpsReadTimeout       = "ops.read_timeout"
	cfgOpsReadHeaderTimeout = "ops.read_header_timeout"
	cfgOpsWriteTimeout      = "ops.write_timeout"
	cfgOpsIdleTimeout       = "ops.idle_timeout"
	cfgOpsMaxHeaderBytes    = "ops.max_header_bytes"
	cfgOpsDisableMetrics    = "ops.disable_metrics"
	cfgOpsDisableProfile    = "ops.disable_profile"
	cfgOpsDisableHealthy    = "ops.disable_healthy"

	opsPathMetrics        = "/metrics"
	opsPathDebugVars      = "/debug/vars"
	opsPathProfileIndex   = "/debug/pprof/"
	opsPathProfileCMDLine = "/debug/pprof/cmdline"
	opsPathProfileProfile = "/debug/pprof/profile"
	opsPathProfileSymbol  = "/debug/pprof/symbol"
	opsPathProfileTrace   = "/debug/pprof/trace"
	opsPathAppReady       = "/-/ready"
	opsPathAppHealthy     = "/-/healthy"
)

var _ = OpsModule

// OpsModule allows import ops http.Server.
// nolint: gochecknoglobals
var OpsModule = module.New(NewOpsServer, dig.Group("services")).AppendConstructor(NewOpsConfig)

// OpsDefaults allows setting default settings for ops server.
func OpsDefaults(v *viper.Viper) {
	v.SetDefault(cfgOpsAddress, opsDefaultAddress)
	v.SetDefault(cfgOpsNetwork, opsDefaultNetwork)

	v.SetDefault(cfgOpsDisableMetrics, false)
	v.SetDefault(cfgOpsDisableProfile, false)
	v.SetDefault(cfgOpsDisableHealthy, false)
}

// PrepareHTTPService creates http.Server as service.Service.
func PrepareHTTPService(cfg HTTPConfig) (service.Service, error) {
	serve := &http.Server{
		Handler: cfg.Handler,

		ReadTimeout:       cfg.ReadTimeout,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
		MaxHeaderBytes:    cfg.MaxHeaderBytes,
	}

	cfg.Logger.Info("creating http.Server",
		zap.String("name", cfg.Name),
		zap.String("address", cfg.Address))

	return NewHTTPService(serve,
		HTTPName(cfg.Name),
		HTTPWithLogger(cfg.Logger),
		HTTPListenAddress(cfg.Address),
		HTTPListenNetwork(cfg.Network))
}

// NewOpsConfig creates OpsConfig and should be moved to settings module in the future.
func NewOpsConfig(v *viper.Viper, l *zap.Logger) (*OpsConfig, error) {
	switch {
	case l == nil:
		return nil, ErrEmptyLogger
	case v == nil:
		return nil, ErrEmptyConfig
	case !v.IsSet(cfgOpsAddress):
		return nil, ErrEmptyHTTPAddress
	}

	hc := HTTPConfig{
		Logger:            l,
		Name:              opsDefaultName,
		Address:           v.GetString(cfgOpsAddress),
		Network:           v.GetString(cfgOpsNetwork),
		ReadTimeout:       v.GetDuration(cfgOpsReadTimeout),
		ReadHeaderTimeout: v.GetDuration(cfgOpsReadHeaderTimeout),
		WriteTimeout:      v.GetDuration(cfgOpsWriteTimeout),
		IdleTimeout:       v.GetDuration(cfgOpsIdleTimeout),
		MaxHeaderBytes:    v.GetInt(cfgOpsMaxHeaderBytes),
	}

	return &OpsConfig{
		HTTPConfig:     hc,
		DisableMetrics: v.GetBool(cfgOpsDisableMetrics),
		DisableProfile: v.GetBool(cfgOpsDisableProfile),
		DisableHealthy: v.GetBool(cfgOpsDisableHealthy),
	}, nil
}

// NewOpsServer creates ops server.
func NewOpsServer(cfg *OpsConfig, probe OpsProbeParams) (service.Service, error) {
	mux := http.NewServeMux()

	mux.Handle("/", http.NotFoundHandler())

	if !cfg.DisableMetrics {
		mux.Handle(opsPathMetrics, promhttp.Handler())
	}

	if !cfg.DisableProfile {
		mux.Handle(opsPathDebugVars, expvar.Handler())

		mux.HandleFunc(opsPathProfileIndex, pprof.Index)
		mux.HandleFunc(opsPathProfileCMDLine, pprof.Cmdline)
		mux.HandleFunc(opsPathProfileProfile, pprof.Profile)
		mux.HandleFunc(opsPathProfileSymbol, pprof.Symbol)
		mux.HandleFunc(opsPathProfileTrace, pprof.Trace)
	}

	if !cfg.DisableHealthy {
		mux.HandleFunc(opsPathAppReady, probeChecker(probe.ReadyProbes))
		mux.HandleFunc(opsPathAppHealthy, probeChecker(probe.HealthProbes))
	}

	return PrepareHTTPService(HTTPConfig{
		Logger:  cfg.Logger,
		Handler: mux,
		Name:    cfg.Name,
		Address: cfg.Address,
		Network: cfg.Network,
	})
}

func probeChecker(probes []ProbeChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		for i := range probes {
			if probes[i] == nil {
				continue
			}

			if err := probes[i](r.Context()); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)

				return
			}
		}

		w.WriteHeader(http.StatusOK)
	}
}
