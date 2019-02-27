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
	}
)

// New helium instance
func New(cfg *Settings, mod module.Module) (*Helium, error) {
	h := &Helium{
		di: dig.New(),
	}

	if cfg != nil {
		if cfg.Prefix == "" {
			cfg.Prefix = cfg.Name
		}

		if tmp := os.Getenv(strings.ToUpper(cfg.Prefix) + "_CONFIG"); tmp != "" {
			cfg.File = tmp
		}

		if tmp := os.Getenv(strings.ToUpper(cfg.Prefix) + "_CONFIG_TYPE"); tmp != "" {
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

		mod = append(mod, core.Provider())
	}

	if err := module.Provide(h.di, mod); err != nil {
		return nil, err
	}

	return h, nil
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
			Name:         "",
			BuildVersion: "",
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
	for {
		var (
			v  = reflect.ValueOf(err)
			fn reflect.Value
		)

		if v.Type().Kind() != reflect.Struct {
			break
		}

		fn = v.FieldByName("Func")
		err = v.FieldByName("Reason").Interface().(error)

		fmt.Printf("Place: %#v\nReason: %s\n\n", fn, err)
	}

	panic(err)
}
