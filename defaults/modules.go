package defaults

import (
	"github.com/im-kulikov/helium/grace"
	"github.com/im-kulikov/helium/logger"
	"github.com/im-kulikov/helium/module"
	"github.com/im-kulikov/helium/orm"
	"github.com/im-kulikov/helium/redis"
	"github.com/im-kulikov/helium/settings"
	"github.com/im-kulikov/helium/web"
	"github.com/im-kulikov/helium/workers"
)

// Module defaults
var Module = module.Module{}.
	Append(
		grace.Module,      // graceful context
		settings.Module,   // settings
		logger.Module,     // logger
		web.ServersModule, // web-servers
		web.EngineModule,  // web-engine
		orm.Module,        // pg-connection
		redis.Module,      // redis-connection
		workers.Module)    // workers
