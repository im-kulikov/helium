package defaults

import (
	"context"

	"github.com/chapsuk/grace"
	"github.com/im-kulikov/helium/logger"
	"github.com/im-kulikov/helium/module"
	"github.com/im-kulikov/helium/orm"
	"github.com/im-kulikov/helium/redis"
	"github.com/im-kulikov/helium/settings"
	"github.com/im-kulikov/helium/web"
)

var Grace = module.Module{
	{Constructor: func() context.Context {
		return grace.ShutdownContext(context.Background())
	}},
}

var Module = module.Module{}.
	Append(Grace).             // graceful context
	Append(settings.Module).   // settings
	Append(logger.Module).     // logger
	Append(web.ServersModule). // web-servers
	Append(orm.Module).        // pg-connection
	Append(redis.Module)       // redis-connection
