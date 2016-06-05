package config

import (
	"encoding/xml"
	"github.com/pfandl/dws/debug"
	"github.com/pfandl/dws/error"
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
	// errors
	NoConfig = "no valid configuration found"
)

type SaneConfig interface {
	IsSane(c *ConfigData) error
}

type IpV4 struct {
	SaneConfig
	XMLName xml.Name `xml:"ipv4"`
	Address string   `xml:"address" validation:"ipv4"`
	Subnet  string   `xml:"subnet"  validation:"ipv4"`
	Mac     string   `xml:"mac"     validation:"mac"`
	Port    string   `xml:"port"    validation:"port"`
}

type Host struct {
	SaneConfig
	XMLName xml.Name `xml:"host"`
	Name    string   `xml:"name,attr" validation:"!empty"`
	IpV4    IpV4     `xml:"ipv4"`
	UtsName string   `xml:"utsname"   validation:"!empty"`
}

type Network struct {
	SaneConfig
	XMLName xml.Name `xml:"network"`
	Name    string   `xml:"name,attr" validation:"!empty,max=15"`
	IpV4    IpV4     `xml:"ipv4"`
	Type    string   `xml:"type"`
	Hosts   []Host   `xml:"host"`
}

type ConfigData struct {
	SaneConfig
	XMLName  xml.Name  `xml:"config"`
	Name     string    `xml:"name,attr" validation:"!empty"`
	Networks []Network `xml:"network"   validation:"slice"`
}

func (d *ConfigData) IsSane(c *ConfigData) error {
	debug.Ver("ConfigData: IsSane")
	return nil
}

func (d *Network) IsSane(c *ConfigData) error {
	debug.Ver("Network: IsSane")
	return nil
}

func (d *Host) IsSane(c *ConfigData) error {
	debug.Ver("Host: IsSane")
	return nil
}

func (d *IpV4) IsSane(c *ConfigData) error {
	debug.Ver("IpV4: IsSane")
	return nil
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
	// read from all paths
	for _, path := range Paths {
		debug.Info("reading config from %s", path)

		// read config file
		data, err := ioutil.ReadFile(path)
		if err != nil {
			debug.Warn("could not read config file from %s (%v)", path, err)
			continue
		}

		// parse xml
		conf := &ConfigData{}
		if err := xml.Unmarshal(data, conf); err != nil {
			debug.Err("could not parse config file %s (%v)", path, err)
			continue
		}

		// validate data
		debug.Info("validating data for %s", conf.Name)
		if err := validation.Validate(*conf, ""); err != nil {
			debug.Fat(err.Error())
		}

		// validate config
		debug.Info("validating config for %s", conf.Name)
		if err := conf.IsSane(conf); err != nil {
			debug.Fat(err.Error())
		}
		return nil
	}
	return err.New(NoConfig)
}

func (c *Config) DisInit() error {
	debug.Ver("Config DisInit()")
	return nil
}

func (c *Config) Event(e string, v interface{}) {
	debug.Ver("Config got event: %s %v", e, v)
}
