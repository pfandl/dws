package module

import (
	"github.com/pfandl/dws/debug"
	"github.com/pfandl/dws/error"
	"github.com/pfandl/dws/event"
)

var (
	Modules = make(map[string]*_module)

	ModuleNameEmpty         = "module name must not be empty"
	ModuleAlreadyRegistered = "module already registered"
	ModuleNotFound          = "module not found"
)

type HasName interface {
	Name() string
}

type HasEvents interface {
	Events(active bool) []string
}

type IsEventListener interface {
	Event(string, interface{})
}

type IsInitiable interface {
	Init() error
}

type IsStartable interface {
	Start() error
}

type IsStoppable interface {
	Stop() error
}

type Module interface {
	HasName
	HasEvents
	IsInitiable
	IsStartable
	IsStoppable
	IsEventListener
}

type _module struct {
	m              Module
	CanHaveAnError error
}

func Get(s string) (Module, error) {
	if Modules[s] != nil {
		return Modules[s].m, nil
	}
	return nil, err.New(ModuleNotFound, s)
}

func Register(m Module) error {
	debug.Ver("Module: Register()")
	s := m.Name()
	debug.Ver("Module: registering %s", s)
	if Modules[s] == nil {
		Modules[s] = &_module{m: m}
		return nil
	}
	return err.New(ModuleAlreadyRegistered, s)
}

func UnRegister(m Module) error {
	s := m.Name()
	if Modules[s] != nil {
		delete(Modules, s)
		return nil
	}
	return err.New(ModuleNotFound, s)
}

func GetError(s string) error {
	if Modules[s] != nil {
		return Modules[s].CanHaveAnError
	}
	return err.New(ModuleNotFound, s)
}

func StartAll() error {
	debug.Ver("Module: InitAll()")
	// active events
	for _, m := range Modules {
		n := m.m.Name()
		debug.Ver("Module: registering events for %s", n)
		for _, e := range m.m.Events(true) {
			debug.Ver("Module: registering event %s", e)
			if _, err := event.RegisterEvent(e); err != nil {
				return err
			}
			debug.Ver("Module: registered event %s for %s", e, n)
		}
	}
	// passive events
	for _, m := range Modules {
		n := m.m.Name()
		debug.Ver("Module: registering callbacks for %s %v", n, m)
		for _, e := range m.m.Events(false) {
			debug.Ver("Module: registering callback %s", e)
			if err := event.RegisterCallback(e, m.m.Event); err != nil {
				return err
			}
			debug.Ver("Module: registered callback %s for %s", e, n)
		}
	}
	// init
	for _, m := range Modules {
		m.CanHaveAnError = m.m.Init()
	}
	// finish all outstanding asynchronous events
	event.Flush()
	// start
	for _, m := range Modules {
		if m.CanHaveAnError == nil {
			m.CanHaveAnError = m.m.Start()
		}
	}
	return nil
}

func StopAll() {
	for _, m := range Modules {
		m.CanHaveAnError = m.m.Stop()
	}
	// passive events
	for _, m := range Modules {
		for _, e := range m.m.Events(false) {
			event.UnRegisterCallback(e, m.m.Event)
		}
	}
	// active events
	for _, m := range Modules {
		for _, e := range m.m.Events(true) {
			event.UnRegisterEvent(e)
		}
	}
}
