package app

import (
	"context"

	"github.com/im-kulikov/helium"
	"github.com/spf13/viper"
	"go.uber.org/dig"
	"go.uber.org/zap"
)

type (
	ServeParams struct {
		dig.In

		Logger *zap.SugaredLogger
		Viper  *viper.Viper
	}

	serveApp struct {
		*zap.SugaredLogger
		*viper.Viper
	}
)

func NewServe(params ServeParams) helium.App {
	return serveApp{
		SugaredLogger: params.Logger,
		Viper:         params.Viper,
	}
}

func (a serveApp) Run(ctx context.Context) error {
	a.Info("app :: serve command")
	for key, val := range a.AllSettings() {
		a.Infof("%s : %#v", key, val)
	}
	return nil
}
