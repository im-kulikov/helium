package app

import (
	"github.com/im-kulikov/helium/grace"
	"github.com/im-kulikov/helium/logger"
	"github.com/im-kulikov/helium/module"
	"github.com/im-kulikov/helium/settings"
)

var (
	mod = module.Module{}.
		Append(
			grace.Module,
			settings.Module,
			logger.Module)

	// ServeCommandModule modules for web-application
	ServeCommandModule = module.New(NewServe).Append(mod)
	// TestCommandModule modules for test
	TestCommandModule = module.New(NewTest).Append(mod)
)
