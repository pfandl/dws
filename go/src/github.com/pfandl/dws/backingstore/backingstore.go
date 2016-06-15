package backingstore

import (
	"github.com/pfandl/dws/config"
	"github.com/pfandl/dws/data"
	"github.com/pfandl/dws/debug"
	"github.com/pfandl/dws/error"
	"github.com/pfandl/dws/event"
	"github.com/pfandl/dws/module"
	"io"
	"net"
	"time"
)

var (
	// events we fire
	ActiveEvents = []string{
		"get-backingstore-path",
		"command-result",
	}
	// events we are interested in
	PassiveEvents = []string{
		"backingstore-available",
		"command",
		"check-command",
	}
	// errors
	BackingStorInvalid = "could not convert backing store"
	// messages
	BackingStoreAdded = "backingstore was added"
	// communication channel
	Channel chan *data.Message
)

type Thread struct {
	Running bool
	Server  interface{}
}

type BackingStore struct {
	module.Module
	Servers []*Thread
}

func (c *BackingStore) Name() string {
	return "backingstore"
}

func (c *BackingStore) Events(active bool) []string {
	debug.Ver("BackingStore: Events %v", active)
	if active == true {
		return ActiveEvents
	} else {
		return PassiveEvents
	}
}

func (c *BackingStore) Event(e string, v interface{}) {
	debug.Ver("BackingStore got event: %s %v", e, v)
	switch e {
	case "backingstore-available":
		c.Available(v)
	case "command":
		c.Command(v.(*data.Message))
	case "check-command":
		c.CheckCommand(v.(*data.Message))
	default:
		debug.Fat("BackingStore event %s unknown", e)
	}
}

func (c *BackingStore) Init() error {
	debug.Ver("BackingStore Init()")
	return nil
}

func (c *BackingStore) Start() error {
	debug.Ver("BackingStore Start()")
	for _, s := range c.Servers {
		if err := s.Start(); err != nil {
			return err
		}
	}
	return nil
}

func (c *Thread) Start() error {
	debug.Ver("Thread Start()")
	if bs, ok := c.Server.(*config.LocalBackingStore); ok {
		if l, err := net.Listen("tcp", ":"+bs.Host.IpV4.Port); err != nil {
			return err
		} else {
			// run in thread
			go c.RunListener(l)
		}
	} else {
		if _, ok := c.Server.(*config.RemoteBackingStore); ok {
			go c.RunTalker()
			return nil
		} else {
			return err.New(BackingStorInvalid)
		}
	}

	Channel = make(chan *data.Message)

	return nil
}

func (c *Thread) RunListener(l net.Listener) {
	debug.Ver("BackingStore RunListener Run()")

	c.Running = true
	for c.Running {
		debug.Ver("BackingStore RunListener Waiting...")
		c, err := l.Accept()
		if err != nil {
			debug.Err("BackingStore RunListener connection failed %s", err.Error())
			continue
		}

		go func(c net.Conn) {
			debug.Ver("BackingStore RunListener connection established with %s", c.LocalAddr().String())

			// close when returning
			defer func() { c.Close() }()

			var msg string
			// should be enough for most messages
			b := make([]byte, 2048)

			for {
				// we are looping to get all data
				s, err := c.Read(b)
				if err != nil && err != io.EOF {
					debug.Warn("BackingStore RunListener listener read failed %s", err.Error())
					c.Close()
					return
				} else {
					// EOF is a welcomed error
					if s <= 0 {
						// stop reading if EOF and no data left to read
						break
					}
					// add part of message
					msg += string(b[0:s])
				}
			}

			// did we receive anything?
			if msg != "" {
				debug.Ver(msg)
			}
		}(c)
	}
}

func (c *Thread) RunTalker() {
	debug.Ver("RemoteBackingStore Thread RunTalker")

	bs := (c.Server).(*config.RemoteBackingStore)

	c.Running = true

	for c.Running {

		// wait till we need to send data
		m := <-Channel

		if l, err := net.Dial("tcp", bs.Host.IpV4.Address+":"+bs.Host.IpV4.Port); err != nil {
			debug.Err(err.Error())
			time.Sleep(10 * time.Second)
			continue
		} else {
			debug.Ver("RemoteBackingStore RunTalker connection established with %s", l.RemoteAddr().String())

			_, err := l.Write([]byte(m.ToJson()))
			if err != nil {
				debug.Err("RemoteBackingStore RunTalker connection failed %s", err.Error())
				continue
			}
			l.Close()
		}
	}
}

func (c *BackingStore) Stop() error {
	debug.Ver("BackingStore Stop()")
	return nil
}

func (c *BackingStore) CheckCommand(m *data.Message) {
	debug.Ver("BackingStore CheckCommand: %v", m)
}

func (c *BackingStore) Command(m *data.Message) {
	debug.Ver("BackingStore Command: %v", m)
	Channel <- m
}

func (c *BackingStore) CreateThread(s interface{}) (*Thread, error) {
	debug.Ver("BackingStore CreateThread: %v", s)
	t := &Thread{
		Running: true,
		Server:  s,
	}
	return t, t.Start()
}

func (c *BackingStore) Add(s interface{}, t *Thread) {
	debug.Ver("BackingStore Add: %v", s)
	if t == nil {
		t = &Thread{
			Server:  s,
			Running: false,
		}
	}
	c.Servers = append(c.Servers, t)
}

func (c *BackingStore) Added(m *data.Message) {
	debug.Ver("BackingStore add: %v", m.Data)
	s := m.Data.(interface{})

	// fire result event after function is done
	defer func() {
		event.Fire("add-backingstore-result", m)
	}()

	if t, err := c.CreateThread(&s); err != nil {
		m.Succeeded = false
		m.Message = err.Error()
	} else {
		m.Succeeded = true
		m.Message = BackingStoreAdded
		c.Add(&s, t)
	}
}

func (c *BackingStore) Available(s interface{}) {
	debug.Ver("BackingStore available: %v", s)
	c.Add(s, nil)
}
