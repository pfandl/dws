package module

import (
	"github.com/pfandl/dws/debug"
	"github.com/pfandl/dws/error"
	"github.com/pfandl/dws/event"
)

var (
	Modules = make(map[string]*Module)

	ModuleNameEmpty         = "module name must not be empty"
	ModuleAlreadyRegistered = "module already registered"
	ModuleNotFound          = "module not found"
)

type _module interface {
	Name() string
	Init() error
	DisInit() error
	Events(active bool) []string
	Event(string, interface{})
}

type Module struct {
	m     *_module
	Error error
}

func Register(m _module) error {
	debug.Ver("Module: Register()")
	s := m.Name()
	debug.Info("Module: registering %s", s)
	if Modules[s] == nil {
		Modules[s] = &Module{&m, nil}
		return nil
	}
	return err.New(ModuleAlreadyRegistered, s)
}

func UnRegister(m _module) error {
	s := m.Name()
	if Modules[s] != nil {
		delete(Modules, s)
		return nil
	}
	return err.New(ModuleNotFound, s)
}

func GetError(s string) error {
	if Modules[s] != nil {
		return Modules[s].Error
	}
	return err.New(ModuleNotFound, s)
}

func InitAll() error {
	debug.Ver("Module: InitAll()")
	// active events
	for _, m := range Modules {
		n := (*m.m).Name()
		debug.Ver("Module: registering events for %s", n)
		for _, e := range (*m.m).Events(true) {
			debug.Ver("Module: registering event %s", e)
			if _, err := event.RegisterEvent(e); err != nil {
				return err
			}
			debug.Info("Module: registered event %s for %s", e, n)
		}
	}
	// passive events
	for _, m := range Modules {
		n := (*m.m).Name()
		debug.Ver("Module: registering callbacks for %s", n)
		for _, e := range (*m.m).Events(false) {
			debug.Ver("Module: registering callback %s", e)
			if err := event.RegisterCallback(e, (*m.m).Event); err != nil {
				return err
			}
			debug.Info("Module: registered callback %s for %s", e, n)
		}
	}
	for _, m := range Modules {
		m.Error = (*m.m).Init()
	}
	return nil
}

func DisInitAll() {
	// passive events
	for _, m := range Modules {
		for _, e := range (*m.m).Events(false) {
			event.UnRegisterCallback(e, (*m.m).Event)
		}
	}
	// active events
	for _, m := range Modules {
		for _, e := range (*m.m).Events(true) {
			event.UnRegisterEvent(e)
		}
	}
	for _, m := range Modules {
		m.Error = (*m.m).DisInit()
	}
}
