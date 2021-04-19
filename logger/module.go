package logger

import (
	"github.com/im-kulikov/helium/module"
)

// Module of loggers
// nolint:gochecknoglobals
var Module = module.Module{
	{Constructor: NewLoggerConfig},
	{Constructor: NewLogger},
	{Constructor: NewStdLogger},
	{Constructor: NewSugaredLogger},
}
