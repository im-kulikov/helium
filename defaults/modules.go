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

var Module = module.Module{}.
	Append(grace.Module).      // graceful context
	Append(settings.Module).   // settings
	Append(logger.Module).     // logger
	Append(web.ServersModule). // web-servers
	Append(orm.Module).        // pg-connection
	Append(redis.Module).      // redis-connection
	Append(workers.Module)     // workers
