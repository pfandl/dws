package event

import (
	"github.com/pfandl/dws/error"
	"time"
)

var (
	Asynchronous = true
	Events       = make(map[string]*ActiveEvent)
	// errors
	EventAlreadyRegistered    = "event already registered"
	EventNotFound             = "event not found"
	CallbackAlreadyRegistered = "callback already registered"
	CallbackNotFound          = "callback not found"
	// next id for asynchronous events
	NextId = uint64(0)
)

type _activeEvent interface {
	Fire(interface{})
	Register(_passiveEvent) error
	RegisterCallback(func(interface{})) error
	UnRegister(_passiveEvent) error
	UnRegisterCallback(func(interface{})) error
}

type _passiveEvent interface {
	Extinguish(interface{})
}

type _event interface {
	_activeEvent
	_passiveEvent
}

type ActiveEvent struct {
	_activeEvent
	Listeners  []_passiveEvent
	Unfinished []uint64
	Callbacks  []*func(string, interface{})
	Name       string
}

func (a *ActiveEvent) Register(p _passiveEvent) error {
	for i := 0; i < len(a.Listeners); i++ {
		if a.Listeners[i] == p {
			return err.New(EventAlreadyRegistered)
		}
	}
	a.Listeners = append(a.Listeners, p)
	return nil
}

func (a *ActiveEvent) UnRegister(p _passiveEvent) error {
	for i := 0; i < len(a.Listeners); i++ {
		if a.Listeners[i] == p {
			a.Listeners = append(a.Listeners[:i], a.Listeners[i+1:]...)
			return nil
		}
	}
	return err.New(EventNotFound)
}

func (a *ActiveEvent) RegisterCallback(c *func(string, interface{})) error {
	for i := 0; i < len(a.Callbacks); i++ {
		if a.Callbacks[i] == c {
			return err.New(CallbackAlreadyRegistered)
		}
	}
	a.Callbacks = append(a.Callbacks, c)
	return nil
}

func (a *ActiveEvent) UnRegisterCallback(c *func(string, interface{})) error {
	for i := 0; i < len(a.Callbacks); i++ {
		if a.Callbacks[i] == c {
			a.Callbacks = append(a.Callbacks[:i], a.Callbacks[i+1:]...)
			return nil
		}
	}
	return err.New(CallbackNotFound)
}

func (a *ActiveEvent) Asynchronous(f func()) {
	// grab next id
	id := NextId
	// and increase it
	NextId++
	// append this id to unfinished list
	a.Unfinished = append(a.Unfinished, id)
	go func(u uint64) {
		// execute passed function
		f()
		// remove this id from unfinished list
		for i := 0; i < len(a.Unfinished); i++ {
			if a.Unfinished[i] == u {
				a.Unfinished = append(a.Unfinished[:i], a.Unfinished[i+1:]...)
				break
			}
		}
	}(id)
}

func (a *ActiveEvent) Fire(v interface{}) {
	for _, l := range a.Listeners {
		// IMPORTANT!!!!
		// cannot use 'l' in asynchronous callback function as it
		// would be overwritten by the next loop and thus only a
		// subset of callbacks (or the last one) will be called!
		ll := l
		if Asynchronous == true {
			a.Asynchronous(func() { ll.Extinguish(v) })
		} else {
			l.Extinguish(v)
		}
	}
	for _, c := range a.Callbacks {
		// IMPORTANT!!!!
		// cannot use 'c' in asynchronous callback function as it
		// would be overwritten by the next loop and thus only a
		// subset of callbacks (or the last one) will be called!
		cc := *c
		if Asynchronous == true {
			a.Asynchronous(func() { cc(a.Name, v) })
		} else {
			cc(a.Name, v)
		}
	}
}

type PassiveEvent struct {
	_passiveEvent
}

func SetAsynchronous(b bool) {
	Asynchronous = b
}

func RegisterEvent(s string) (*ActiveEvent, error) {
	if Events[s] == nil {
		a := &ActiveEvent{Name: s}
		Events[s] = a
		return a, nil
	}
	return Events[s], nil
}

func UnRegisterEvent(s string) error {
	if Events[s] != nil {
		delete(Events, s)
		return nil
	}
	return err.New(EventNotFound, s)
}

func RegisterListener(s string, p _passiveEvent) error {
	if Events[s] != nil {
		return Events[s].Register(p)
	}
	return err.New(EventNotFound, s)
}

func UnRegisterListener(s string, p _passiveEvent) error {
	if Events[s] != nil {
		return Events[s].UnRegister(p)
	}
	return err.New(EventNotFound, s)
}

func RegisterCallback(s string, c func(string, interface{})) error {
	if Events[s] != nil {
		return Events[s].RegisterCallback(&c)
	}
	return err.New(EventNotFound, s)
}

func UnRegisterCallback(s string, c func(string, interface{})) error {
	if Events[s] != nil {
		return Events[s].UnRegisterCallback(&c)
	}
	return err.New(EventNotFound, s)
}

func Fire(s string, v interface{}) error {
	if Events[s] != nil {
		Events[s].Fire(v)
		return nil
	}
	return err.New(EventNotFound, s)
}

func Flush() {
	for _, e := range Events {
		for len(e.Unfinished) > 0 {
			time.Sleep(1 * time.Millisecond)
		}
	}
}
