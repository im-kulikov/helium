package grace

import (
	"github.com/im-kulikov/helium/module"
)

// Module graceful context.
// nolint:gochecknoglobals
var Module = module.Module{
	{Constructor: NewGracefulContext},
}
