package server

import (
	"github.com/pfandl/dws/debug"
	"github.com/pfandl/dws/module"
)

var (
	// events we fire
	ActiveEvents = []string{
		"add-server",
		"remove-server",
		"add-backingstore",
		"remove-backingstore",
		"add-network",
		"remove-network",
		"add-host",
		"remove-host",
	}
	// events we are interested in
	PassiveEvents = []string{
		"add-server-result",
		"remove-server-result",
		"add-backingstore-result",
		"remove-backingstore-result",
		"add-network-result",
		"remove-network-result",
		"add-host-result",
		"remove-host-result",
	}
)

type Server struct {
	module.Module
}

func (c *Server) Name() string {
	return "server"
}

func (c *Server) Events(active bool) []string {
	debug.Ver("Server: Events %v", active)
	if active == true {
		return ActiveEvents
	} else {
		return PassiveEvents
	}
}

func (c *Server) Init() error {
	debug.Ver("Server Init()")
	return nil
}

func (c *Server) DisInit() error {
	debug.Ver("Server DisInit()")
	return nil
}

func (c *Server) Event(e string, v interface{}) {
	debug.Ver("Server got event: %s %v", e, v)
}