package app

import (
	"context"

	"github.com/im-kulikov/helium"
	"github.com/spf13/viper"
	"go.uber.org/dig"
	"go.uber.org/zap"
)

type (
	TestParams struct {
		dig.In

		Logger *zap.SugaredLogger
		Viper  *viper.Viper
	}

	testApp struct {
		*zap.SugaredLogger
		*viper.Viper
	}
)

func NewTest(params ServeParams) helium.App {
	return testApp{
		SugaredLogger: params.Logger,
		Viper:         params.Viper,
	}
}

func (a testApp) Run(ctx context.Context) error {
	a.Info("app :: test command")
	for key, val := range a.AllSettings() {
		a.Infof("%s : %#v", key, val)
	}
	return nil
}
