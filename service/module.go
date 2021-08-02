package service

import (
	"time"

	"github.com/spf13/viper"
	"go.uber.org/dig"

	"github.com/im-kulikov/helium/module"
)

type outParams struct {
	dig.Out

	Shutdown time.Duration `name:"service_shutdown_timeout"`
}

// ShutdownTimeoutParam name for viper setting.
const ShutdownTimeoutParam = "shutdown_timeout"

var (
	_ = Module // prevent unused

	// Module for group of services
	// nolint:gochecknoglobals
	Module = module.Module{
		{Constructor: newParam},
		{Constructor: newGroup},
	}
)

func newParam(v *viper.Viper) outParams {
	return outParams{Shutdown: v.GetDuration(ShutdownTimeoutParam)}
}
