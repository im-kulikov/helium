package grace

import (
	"github.com/im-kulikov/helium/module"
)

// Module graceful context
var Module = module.Module{
	{Constructor: NewGracefulContext},
}
