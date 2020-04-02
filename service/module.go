package service

import "github.com/im-kulikov/helium/module"

var (
	_ = Module // prevent unused

	// Module for group of services
	Module = module.New(newGroup)
)
