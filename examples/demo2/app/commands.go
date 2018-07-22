package app

import (
	"github.com/im-kulikov/helium/defaults"
	"github.com/im-kulikov/helium/logger"
	"github.com/im-kulikov/helium/module"
	"github.com/im-kulikov/helium/settings"
)

var (
	mod = module.Module{}.
		Append(defaults.Grace).
		Append(settings.Module).
		Append(logger.Module)

	ServeCommandModule = module.Module{
		{Constructor: NewServe},
	}.Append(mod)

	TestCommandModule = module.Module{
		{Constructor: NewTest},
	}.Append(mod)
)
