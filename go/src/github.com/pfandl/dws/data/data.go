package data

import (
	"encoding/json"
	"github.com/pfandl/dws/debug"
)

var (
	CannotConvertToJson = "cannot convert data to json"
)

type Encapsulation interface {
	ToJson() string
	FromJson(string) error
}

type Message struct {
	Encapsulation
	Succeeded bool
	Message   string
	Data      interface{}
}

func (m *Message) ToJson() string {
	debug.Ver("Message: ToJson %v", m)
	if d, e := json.Marshal(m); e != nil {
		return "{Succeeded: false, Message: " + CannotConvertToJson + "}"
	} else {
		return string(d)
	}
}

func (m *Message) FromJson(s string) error {
	if e := json.Unmarshal([]byte(s), m); e != nil {
		return e
	}
	return nil
}

func ToJson(b bool, s string, d interface{}) string {
	m := Message{
		Succeeded: b,
		Message:   s,
		Data:      d,
	}
	return m.ToJson()
}

func FromJson(s string, d interface{}) error {
	m := d.(*Message)
	return m.FromJson(s)
}
