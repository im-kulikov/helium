package helium

import (
	"context"
	"fmt"
	stdlog "log"
	"os"
	"reflect"
	"strings"

	"github.com/im-kulikov/helium/logger"
	"github.com/im-kulikov/helium/module"
	"github.com/im-kulikov/helium/settings"
	"github.com/spf13/viper"
	"go.uber.org/atomic"
	"go.uber.org/dig"
	"go.uber.org/zap"
)

type (
	// App implementation for helium
	App interface {
		Run(ctx context.Context) error
	}

	// Helium struct
	Helium struct {
		di *dig.Container
	}

	// Settings struct
	Settings struct {
		File         string
		Type         string
		Name         string
		Prefix       string
		BuildTime    string
		BuildVersion string
		Defaults     settings.Defaults
	}
)

var (
	appName    = atomic.NewString("helium")
	appVersion = atomic.NewString("dev")
)

// New helium instance
func New(cfg *Settings, mod ...module.Module) (*Helium, error) {
	h := &Helium{di: dig.New()}

	modules := module.Combine(mod...)

	if cfg != nil {
		if cfg.Prefix == "" {
			cfg.Prefix = cfg.Name
		}
		cfg.Prefix = strings.ToUpper(cfg.Prefix)

		if tmp := os.Getenv(cfg.Prefix + "_CONFIG"); tmp != "" {
			cfg.File = tmp
		}

		if tmp := os.Getenv(cfg.Prefix + "_CONFIG_TYPE"); tmp != "" {
			cfg.Type = tmp
		}

		core := settings.Core{
			File:         cfg.File,
			Type:         cfg.Type,
			Name:         cfg.Name,
			Prefix:       cfg.Prefix,
			BuildTime:    cfg.BuildTime,
			BuildVersion: cfg.BuildVersion,
		}

		appName.Store(cfg.Name)
		appVersion.Store(cfg.BuildVersion)

		modules = append(modules, core.Provider())
		modules = append(modules, settings.DIProvider(h.di))
	}

	if err := module.Provide(h.di, modules); err != nil {
		return nil, err
	}

	if cfg == nil || cfg.Defaults == nil {
		return h, nil
	}

	return h, h.di.Invoke(cfg.Defaults)
}

// Invoke dependencies from DI container
func (h Helium) Invoke(fn interface{}, args ...dig.InvokeOption) error {
	return h.di.Invoke(fn, args...)
}

// Run trying invoke app instance from DI container and start app with Run call
func (h Helium) Run() error {
	return h.di.Invoke(func(ctx context.Context, app App) error {
		return app.Run(ctx)
	})
}

// Catch errors
func Catch(err error) {
	if err == nil {
		return
	}

	v := viper.New()
	log, logErr := logger.
		NewLogger(logger.NewLoggerConfig(v), &settings.Core{
			Name:         appName.Load(),
			BuildVersion: appVersion.Load(),
		})
	if logErr != nil {
		stdlog.Fatal(err)
	} else {
		log.Fatal("Can't run app",
			zap.Error(err))
	}
}

// CatchTrace catch errors for debugging
// use that function just for debug your application
func CatchTrace(err error) {
	if err == nil {
		return
	}

	// digging into the root of the problem
loop:
	for {
		var (
			ok bool
			v  = reflect.ValueOf(err)
			fn reflect.Value
		)

		switch {
		case v.Type().Kind() != reflect.Struct,
			!v.FieldByName("Reason").IsValid():
			break loop
		case v.FieldByName("Func").IsValid():
			fn = v.FieldByName("Func")
		}

		fmt.Printf("Place: %#v\nReason: %s\n\n", fn, err)

		if err, ok = v.FieldByName("Reason").Interface().(error); !ok {
			err = v.Interface().(error)
			break
		}
	}

	panic(err)
}
