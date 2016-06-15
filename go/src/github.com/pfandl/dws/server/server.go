package server

import (
	"github.com/pfandl/dws/config"
	"github.com/pfandl/dws/data"
	"github.com/pfandl/dws/debug"
	"github.com/pfandl/dws/event"
	"github.com/pfandl/dws/module"
	"net"
)

var (
	// events we fire
	ActiveEvents = []string{
		"command",
	}
	// events we are interested in
	PassiveEvents = []string{
		"server-available",
		"command-result",
		"check-command",
	}
	// errors
	// messages
	ServerAdded = "server was added"
)

type Thread struct {
	Running bool
	Server  *config.Server
}

type Server struct {
	module.Module
	Servers []*Thread
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

func (c *Server) Start() error {
	debug.Ver("Server Start()")
	for _, s := range c.Servers {
		if err := s.Start(); err != nil {
			return err
		}
	}
	return nil
}

func (c *Thread) Start() error {
	debug.Ver("Thread Start()")
	if l, err := net.Listen("tcp", ":"+c.Server.IpV4.Port); err != nil {
		return err
	} else {
		// run in thread
		go c.Run(&l)
	}
	return nil
}

func (c *Thread) Run(l *net.Listener) {
	debug.Ver("Thread Run()")

	// should be enough for most messages
	s := 2048
	b := make([]byte, s)

	c.Running = true
	for c.Running {
		debug.Ver("Thread Waiting...()")
		c, err := (*l).Accept()
		if err != nil {
			debug.Err("Thread connection failed %s", err.Error())
			continue
		}
		debug.Ver("Thread connection established wtih %s", c.RemoteAddr().String())
		c.Read(b)
		c.Close()
	}
}

func (c *Server) Stop() error {
	debug.Ver("Server Stop()")
	return nil
}

func (c *Server) Event(e string, v interface{}) {
	debug.Ver("Server got event: %s %v", e, v)
	switch e {
	case "server-available":
		c.Available(v.(*config.Server))
	case "check-server-added":
		c.Added(v.(*data.Message))
	case "add-host-result":
		fallthrough
	case "add-network-result":
		fallthrough
	case "add-server-result":
		c.Result(v.(*data.Message))
	default:
		debug.Fat("Server event %s unknown", e)
	}
}

func (c *Server) Result(m *data.Message) {
	debug.Ver("Server Result: %v", m)
}

func (c *Server) CreateThread(s *config.Server) (*Thread, error) {
	debug.Ver("Server CreateThread: %v", s)
	t := &Thread{
		Running: true,
		Server:  s,
	}
	return t, t.Start()
}

func (c *Server) Add(s *config.Server, t *Thread) {
	debug.Ver("Server Add: %v", s)
	if t == nil {
		t = &Thread{
			Server:  s,
			Running: false,
		}
	}
	c.Servers = append(c.Servers, t)
}

func (c *Server) Added(m *data.Message) {
	debug.Ver("Server add: %v", m.Data)
	s := m.Data.(config.Server)

	// fire result event after function is done
	defer func() {
		event.Fire("add-server-result", m)
	}()

	if t, err := c.CreateThread(&s); err != nil {
		m.Succeeded = false
		m.Message = err.Error()
	} else {
		m.Succeeded = true
		m.Message = ServerAdded
		c.Add(&s, t)
	}
}

func (c *Server) Available(s *config.Server) {
	debug.Ver("Server available: %v", s)
	c.Add(s, nil)
}
