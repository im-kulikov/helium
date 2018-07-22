package logger

import (
	"github.com/im-kulikov/helium/module"
)

var Module = module.Module{
	{Constructor: NewLoggerConfig},
	{Constructor: NewLogger},
	{Constructor: NewStdLogger},
	{Constructor: NewSugaredLogger},
}
