package config

import (
	"encoding/xml"
	"github.com/pfandl/dws/debug"
	"github.com/pfandl/dws/module"
	"github.com/pfandl/dws/validation"
	"io/ioutil"
)

var (
	// events we fire
	ActiveEvents = []string{
		"server-added",
		"server-removed",
		"backingstore-added",
		"backingstore-removed",
		"network-added",
		"network-removed",
		"host-added",
		"host-removed",
	}
	// events we are interested in
	PassiveEvents = []string{
		"add-server",
		"remove-server",
		"add-backingstore",
		"remove-backingstore",
		"add-network",
		"remove-network",
		"add-host",
		"remove-host",
	}
	// Paths to gather config information ascending in importance
	Paths = []string{
		"./config",
		"/etc/dws/config",
		"/etc/default/dws",
	}
)

type ConfigData struct {
	XMLName xml.Name `xml:"config"`
	Name    string   `xml:"name,attr" validation:"!empty,max=15"`
}

type Config struct {
	module.Module
}

func (c *Config) Name() string {
	return "config"
}

func (c *Config) Events(active bool) []string {
	debug.Ver("Config: Events %v", active)
	if active == true {
		return ActiveEvents
	} else {
		return PassiveEvents
	}
}

func (c *Config) Init() error {
	debug.Ver("Config Init()")
	for _, path := range Paths {
		debug.Info("reading config from %s", path)
		data, err := ioutil.ReadFile(path)
		if err != nil {
			debug.Warn("could not read config file from %s (%v)", path, err)
		} else {
			conf := ConfigData{}
			err = xml.Unmarshal(data, &conf)
			if err != nil {
				debug.Err("could not parse config file %s (%v)", path, err)
			} else {
				debug.Info("read config for %s", conf.Name)
				if err := validation.Validate(conf, ""); err != nil {
					debug.Fat(err.Error())
				}
			}
		}
	}
	return nil
}

func (c *Config) DisInit() error {
	debug.Ver("Config DisInit()")
	return nil
}

func (c *Config) Event(e string, v interface{}) {
	debug.Ver("Config got event: %s %v", e, v)
}
