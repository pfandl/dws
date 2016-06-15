package config

import (
	"encoding/xml"
	"github.com/pfandl/dws/data"
	"github.com/pfandl/dws/debug"
	"github.com/pfandl/dws/error"
	"github.com/pfandl/dws/event"
	"github.com/pfandl/dws/module"
	"github.com/pfandl/dws/validation"
	"io/ioutil"
	"reflect"
	"strings"
)

const (
	Production = 1
	Backup
	Temporary
)

const (
	Btrfs = 1
)

var (
	// events we fire
	ActiveEvents = []string{
		// events fired after config reading
		"server-available",
		"backingstore-available",
		"network-available",
		"host-available",
		// events fired after executing commands
		// when we return the result to server
		"command-result",
		// events fired after executing commands
		// when we are ok and other modules return the result to server
		"check-command",
	}
	// events we are interested in
	PassiveEvents = []string{
		"command",
	}
	// Paths to gather config information ascending in importance
	Paths = []string{
		"./config",
		"/etc/dws/config",
		"/etc/default/dws",
	}
	// network types
	Networks = map[string]int{
		"production": Production,
		"backup":     Backup,
		"temporary":  Temporary,
	}
	// backing store types
	BackingStores = map[string]int{
		"btrfs": Btrfs,
	}
	// our config data
	LoadedConfig *Config
	// errors
	NoConfig                    = "no valid configuration found"
	BackingStoreNameAlreadyUsed = "backing store name is already used"
	BackingStorePortAlreadyUsed = "backing store port is already used"
	ServerNameAlreadyUsed       = "server name is already used"
	NetworkNameAlreadyUsed      = "network name is already used"
	InvalidNetworkType          = "network type is invalid"
	IpV4Overlap                 = "network ip/subnet is overlapping with existing network"
	IpV4AlreadyUsed             = "ip is already in use"
	IpV4Unavailable             = "ip address unavailable"
	HostNameAlreadyUsed         = "host name is already in use"
	UtsNameAlreadyUsed          = "uts name is already in use"
	WrongSubnet                 = "subnets do not match"
	ServerNotFound              = "server not found"
	NetworkNotFound             = "network not found"
	// messages
	ServerAdded  = "server was added"
	HostAdded    = "host was added"
	NetworkAdded = "network was added"
)

type Propagate interface {
	Available() error
}

type SaneConfig interface {
	IsSane(c *ConfigData, s string) error
}

type IpV4 struct {
	SaneConfig
	XMLName xml.Name `xml:"ipv4"`
	Address string   `xml:"address" validation:"ipv4"`
	Subnet  string   `xml:"subnet"  validation:"ipv4"`
	Mac     string   `xml:"mac"     validation:"ipv4mac"`
	Port    string   `xml:"port"    validation:"port"`
}

type ServerIpV4 struct {
	IpV4    `validation:"struct" validation-ignore:"address,subnet,mac"`
	XMLName xml.Name `xml:"ipv4"`
}

type NetworkIpV4 struct {
	IpV4    `validation:"struct" validation-ignore:"mac,port"`
	XMLName xml.Name `xml:"ipv4"`
}

type LocalBackingStoreHostIpV4 struct {
	IpV4    `validation:"struct" validation-ignore:"address,subnet,mac"`
	XMLName xml.Name `xml:"ipv4"`
}

type RemoteBackingStoreHostIpV4 struct {
	IpV4    `validation:"struct" validation-ignore:"subnet,mac"`
	XMLName xml.Name `xml:"ipv4"`
}

type HostIpV4 struct {
	IpV4    `validation:"struct" validation-ignore:"subnet,port"`
	XMLName xml.Name `xml:"ipv4"`
}

type Host struct {
	Propagate
	SaneConfig
	XMLName xml.Name `xml:"host"`
	Name    string   `xml:"name,attr" validation:"!empty"`
	IpV4    HostIpV4 `xml:"ipv4"      validation:"struct"`
	UtsName string   `xml:"utsname"   validation:"uts"`
	Network string
}

type LocalBackingStoreHost struct {
	Host
	XMLName xml.Name                  `xml:"host"`
	IpV4    LocalBackingStoreHostIpV4 `xml:"ipv4"  validation:"struct"`
}

type RemoteBackingStoreHost struct {
	Host
	XMLName xml.Name                   `xml:"host"`
	IpV4    RemoteBackingStoreHostIpV4 `xml:"ipv4"  validation:"struct"`
}

type Network struct {
	Propagate
	SaneConfig
	XMLName xml.Name    `xml:"network"`
	Name    string      `xml:"name,attr" validation:"!empty,max=15"`
	IpV4    NetworkIpV4 `xml:"ipv4"      validation:"struct"`
	Type    string      `xml:"type"      validation:"!empty"`
	Hosts   []Host      `xml:"host"      validation:"slice"`
	Server  string
}

type BackingStore interface {
	Propagate
	SaneConfig
}

type LocalBackingStore struct {
	BackingStore
	XMLName xml.Name              `xml:"backingstore"`
	Name    string                `xml:"name,attr"     validation:"!empty"`
	Host    LocalBackingStoreHost `xml:"host"          validation:"struct"`
}

type RemoteBackingStore struct {
	BackingStore
	XMLName xml.Name               `xml:"backingstore"`
	Host    RemoteBackingStoreHost `xml:"host"          validation:"struct"`
}

type Log struct {
	Propagate
	SaneConfig
	XMLName xml.Name `xml:"log"`
}

type Server struct {
	Propagate
	SaneConfig
	XMLName      xml.Name           `xml:"server"`
	Name         string             `xml:"name,attr"    validation:"!empty"`
	IpV4         ServerIpV4         `xml:"ipv4"         validation:"struct"`
	BackingStore RemoteBackingStore `xml:"backingstore" validation:"struct"`
	Networks     []Network          `xml:"network"      validation:"slice"`
	Log          Log                `xml:"log"`
}

type ConfigData struct {
	Propagate
	SaneConfig
	XMLName       xml.Name            `xml:"config"`
	Name          string              `xml:"name,attr"    validation:"!empty"`
	Servers       []Server            `xml:"server"       validation:"slice"`
	BackingStores []LocalBackingStore `xml:"backingstore" validation:"slice"`
	Validate      bool
}

func (d *ConfigData) Available() error {
	debug.Ver("ConfigData: Available")

	for i := 0; i < len(d.Servers); i++ {
		if err := d.Servers[i].Available(); err != nil {
			return err
		}
	}
	for i := 0; i < len(d.BackingStores); i++ {
		if err := d.BackingStores[i].Available(); err != nil {
			return err
		}
	}

	return nil
}

func (d *ConfigData) IsSane(c *ConfigData, s string) error {
	debug.Ver("ConfigData: IsSane")

	// set to true after conf is read, this is to prevent
	// double validation of data on config read
	if c.Validate == true {
		if err := validation.Validate(*d, s, ""); err != nil {
			return err
		}
	}

	for i := 0; i < len(c.Servers); i++ {
		srv := &c.Servers[i]
		if err := srv.IsSane(c, s); err != nil {
			return err
		}
	}
	for i := 0; i < len(c.BackingStores); i++ {
		srv := &c.BackingStores[i]
		if err := srv.IsSane(c, s); err != nil {
			return err
		}
	}

	return nil
}

func (d *LocalBackingStore) Available() error {
	debug.Ver("BackingStore: Available")
	event.Fire("backingstore-available", d)
	return nil
}

func (d *RemoteBackingStore) Available() error {
	debug.Ver("BackingStore: Available")
	event.Fire("backingstore-available", d)
	return nil
}

func (d *LocalBackingStore) IsSane(c *ConfigData, s string) error {
	debug.Ver("BackingStore: IsSane")

	// set to true after conf is read, this is to prevent
	// double validation of data on config read
	if c.Validate == true {
		if err := validation.Validate(*d, s, ""); err != nil {
			return err
		}
	}

	debug.Ver("BackingStore: IsSane checking host %v", d.Host)
	if err := d.Host.IsSane(c, s); err != nil {
		return err
	}

	for i := 0; i < len(c.BackingStores); i++ {
		s := &c.BackingStores[i]
		// skip same object (do not compare to itself)
		if s == d {
			continue
		}
		debug.Ver("BackingStore: IsSane checking names %s %s", d.Name, s.Name)
		if d.Name == s.Name {
			return err.New(BackingStoreNameAlreadyUsed, d.Name)
		}
	}

	return nil
}

func (d *RemoteBackingStore) IsSane(c *ConfigData, s string) error {
	debug.Ver("BackingStore: IsSane")

	// set to true after conf is read, this is to prevent
	// double validation of data on config read
	if c.Validate == true {
		if err := validation.Validate(*d, s, ""); err != nil {
			return err
		}
	}

	return nil
}

func (d *LocalBackingStoreHost) IsSane(c *ConfigData, s string) error {
	debug.Ver("LocalBackingStoreHost: IsSane")

	// set to true after conf is read, this is to prevent
	// double validation of data on config read
	if c.Validate == true {
		if err := validation.Validate(*d, s, ""); err != nil {
			return err
		}
	}

	debug.Ver("LocalBackingStoreHost: IsSane checking ipv4 %v", d.IpV4)
	if err := d.IpV4.IsSane(c, s); err != nil {
		return err
	}

	return nil
}

func (d *LocalBackingStoreHostIpV4) IsSane(c *ConfigData, s string) error {
	debug.Ver("LocalBackingStoreHostIpV4: IsSane %v", &d)

	// set to true after conf is read, this is to prevent
	// double validation of data on config read
	if c.Validate == true {
		if err := validation.Validate(*d, s, ""); err != nil {
			return err
		}
	}

	for i := 0; i < len(c.BackingStores); i++ {
		s := &c.BackingStores[i].Host.IpV4
		// skip same object (do not compare to itself)
		if s == d {
			continue
		}
		debug.Ver("LocalBackingStoreHostIpV4: IsSane checking port %s %s", d.Port, s.Port)
		if d.Port == s.Port {
			return err.New(BackingStorePortAlreadyUsed, d.Port)
		}
	}

	return nil
}

func (d *Server) Available() error {
	debug.Ver("Server: Available")
	event.Fire("server-available", d)
	for i := 0; i < len(d.Networks); i++ {
		if err := d.Networks[i].Available(); err != nil {
			return err
		}
	}
	d.BackingStore.Available()
	return nil
}

func (d *Server) IsSane(c *ConfigData, s string) error {
	debug.Ver("Server: IsSane")

	// set to true after conf is read, this is to prevent
	// double validation of data on config read
	if c.Validate == true {
		if err := validation.Validate(*d, s, ""); err != nil {
			return err
		}
	}

	for i := 0; i < len(c.Servers); i++ {
		s := &c.Servers[i]
		// skip same object (do not compare to itself)
		if s == d {
			continue
		}
		debug.Ver("Server: IsSane checking names %s %s", d.Name, s.Name)
		if d.Name == s.Name {
			return err.New(ServerNameAlreadyUsed, d.Name)
		}
	}

	for i := 0; i < len(d.Networks); i++ {
		n := &d.Networks[i]
		if err := n.IsSane(c, s); err != nil {
			return err
		}
	}

	return nil
}

func (d *Network) Available() error {
	debug.Ver("Network: Available")
	event.Fire("network-available", d)
	for i := 0; i < len(d.Hosts); i++ {
		if err := d.Hosts[i].Available(); err != nil {
			return err
		}
	}
	return nil
}

func (d *Network) IsSane(c *ConfigData, s string) error {
	debug.Ver("Network: IsSane")

	// set to true after conf is read, this is to prevent
	// double validation of data on config read
	if c.Validate == true {
		if err := validation.Validate(*d, s, ""); err != nil {
			return err
		}
	}

	if Networks[d.Type] == 0 {
		return err.New(InvalidNetworkType, d.Type)
	}

	if !reflect.ValueOf(d.IpV4.IpV4).CanAddr() {
		return err.New(IpV4Unavailable, d.Name)
	}
	debug.Ver("Network: IsSane checking ip %v", d.IpV4)
	if err := d.IpV4.IsSane(c, s); err != nil {
		return err
	}
	// check network names and hosts
	for i := 0; i < len(c.Servers); i++ {
		srv := &c.Servers[i]
		for i := 0; i < len(srv.Networks); i++ {
			n := &srv.Networks[i]
			// skip same object (do not compare to itself)
			if d == n {
				continue
			}
			debug.Ver("Network: IsSane checking names %s %s", d.Name, n.Name)
			if d.Name == n.Name {
				return err.New(NetworkNameAlreadyUsed, d.Name)
			}
			for i := 0; i < len(n.Hosts); i++ {
				h := &n.Hosts[i]
				if err := h.IsSane(c, s); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (d *Host) Available() error {
	debug.Ver("Host: Available")
	event.Fire("host-available", d)
	return nil
}

func (d *Host) IsSane(c *ConfigData, s string) error {
	debug.Ver("Host: IsSane")

	// set to true after conf is read, this is to prevent
	// double validation of data on config read
	if c.Validate == true {
		if err := validation.Validate(*d, s, ""); err != nil {
			return err
		}
	}

	nf := false
	for i := 0; i < len(c.Servers); i++ {
		srv := &c.Servers[i]
		for i := 0; i < len(srv.Networks); i++ {
			n := &srv.Networks[i]
			for i := 0; i < len(n.Hosts); i++ {
				h := &n.Hosts[i]
				// skip same object (do not compare to itself)
				if d == h {
					continue
				}
				debug.Ver("Host: IsSane checking names %s %s", d.Name, h.Name)
				// check name
				if d.Name == h.Name {
					return err.New(HostNameAlreadyUsed, d.Name)
				}
				debug.Ver("Host: IsSane checking uts names %s %s", d.UtsName, h.UtsName)
				// check uts name
				if d.UtsName == h.UtsName {
					return err.New(UtsNameAlreadyUsed, d.UtsName)
				}
				// check ip
				if err := d.IpV4.IsSane(c, s); err != nil {
					return err
				}
			}
			// if we get an add-host event, we have to check if the
			// passed network is actually set up
			if d.Network != "" && n.Name == d.Network {
				nf = true
			}
		}
	}

	// if we get an add-host event, we have to check if the
	// passed network is actually set up
	if d.Network != "" && nf == false {
		return err.New(NetworkNotFound, d.Network)
	}

	return nil
}

func (d *NetworkIpV4) IsSane(c *ConfigData, s string) error {
	debug.Ver("NetworkIpV4: IsSane")

	// set to true after conf is read, this is to prevent
	// double validation of data on config read
	if c.Validate == true {
		if err := validation.Validate(*d, s, ""); err != nil {
			return err
		}
	}

	for i := 0; i < len(c.Servers); i++ {
		srv := &c.Servers[i]
		for i := 0; i < len(srv.Networks); i++ {
			n := &srv.Networks[i]
			// skip same object (do not compare to itself)
			if d == &n.IpV4 {
				continue
			}
			debug.Ver("NetworkIpV4: IsSane checking network ip %s %s", d.Address, n.IpV4.Address)
			if d.Address == n.IpV4.Address {
				return err.New(IpV4AlreadyUsed, d.Address)
			}
			debug.Ver("NetworkIpV4: IsSane checking network subnet %s", d.Subnet)
			// check if network has same ip range as existing
			// TODO: implement the real logic
			match := 0
			ip1 := strings.Split(d.Address, ".")
			ip2 := strings.Split(n.IpV4.Address, ".")
			sn := strings.Split(d.Subnet, ".")
			var l int
			for i := 0; i < 4; i++ {
				if sn[i] != "255" {
					l = i
					break
				}
			}
			for i := 0; i < l; i++ {
				if ip1[i] != ip2[i] {
					break
				}
				match++
			}
			if match >= l && match != 0 {
				return err.New(IpV4Overlap, n.Name)
			}
		}
	}

	return nil
}

func (d *HostIpV4) IsSane(c *ConfigData, s string) error {
	debug.Ver("HostIpV4: IsSane")

	// set to true after conf is read, this is to prevent
	// double validation of data on config read
	if c.Validate == true {
		if err := validation.Validate(*d, s, ""); err != nil {
			return err
		}
	}

	for i := 0; i < len(c.Servers); i++ {
		srv := &c.Servers[i]
		for i := 0; i < len(srv.Networks); i++ {
			n := &srv.Networks[i]
			for i := 0; i < len(n.Hosts); i++ {
				h := &n.Hosts[i]
				// skip same object (do not compare to itself)
				if d == &h.IpV4 {
					continue
				}
				debug.Ver("HostIpV4: IsSane checking host ip %s %s", d.Address, h.IpV4.Address)
				if d.Address == h.IpV4.Address {
					return err.New(IpV4AlreadyUsed, d.Address)
				}
			}
		}
	}

	return nil
}

type Config struct {
	module.Module
	Data *ConfigData
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

func (c *Config) Event(e string, v interface{}) {
	debug.Ver("Config got event: %s %v", e, v)
	switch e {
	case "command":
		c.Command(v.(*data.Message))
	default:
		debug.Fat("Config event %s unknown", e)
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

		// set ValidateData to false so that we dont validate in the IsSane
		// functions again, we will check data once ourselves here instead
		conf := &ConfigData{Validate: false}

		// parse xml
		if err := xml.Unmarshal(data, conf); err != nil {
			debug.Err("could not parse config file %s (%v)", path, err)
			continue
		}

		// validate data
		debug.Info("validating data for %s", conf.Name)
		if err := validation.Validate(*conf, "", ""); err != nil {
			debug.Fat(err.Error())
		}

		// validate config
		debug.Info("validating config for %s", conf.Name)
		if err := conf.IsSane(conf, ""); err != nil {
			debug.Fat(err.Error())
		}

		// use this config
		c.Data = conf
		// validate data in IsSane
		c.Data.Validate = true

		return c.Data.Available()
	}
	return err.New(NoConfig)
}

func (c *Config) Start() error {
	debug.Ver("Config Start()")
	return nil
}

func (c *Config) Stop() error {
	debug.Ver("Config Stop()")
	return nil
}

func (c *Config) Command(m *data.Message) {
	debug.Ver("Config Command: %v", m)
	switch m.Message {
	case "add-server":
		c.AddServer(m)
	case "add-network":
		c.AddNetwork(m)
	case "add-host":
		c.AddHost(m)
	}
}

func AfterCommand(m *data.Message) {
	if m.Succeeded == true {
		// host module should also check data
		event.Fire("check-command", m)
	} else {
		// something went wrong, inform server about that
		event.Fire("command-result", m)
	}
}

func (c *Config) AddServer(m *data.Message) {
	debug.Ver("Config AddServer: %v", m)
	s := m.Data.(Server)

	// fire result event after function is done
	defer AfterCommand(m)

	if e := s.IsSane(c.Data, "address,subnet,mac"); e != nil {
		// something is wrong with specified data
		m.Message = e.Error()
		m.Succeeded = false
	} else {
		c.Data.Servers = append(c.Data.Servers, s)
		m.Succeeded = true
	}
}

func (c *Config) AddNetwork(m *data.Message) {
	debug.Ver("Config AddNetwork: %v", m)
	n := m.Data.(Network)

	// fire result event after function is done
	defer AfterCommand(m)

	if e := n.IsSane(c.Data, "mac,port"); e != nil {
		// something is wrong with specified data
		m.Message = e.Error()
		m.Succeeded = false
	} else {
		for _, s := range c.Data.Servers {
			if s.Name == n.Server {
				s.Networks = append(s.Networks, n)
				m.Succeeded = true
				return
			}
		}
		// uhm... no network with this name found
		m.Succeeded = false
		m.Message = err.New(ServerNotFound, n.Server).Error()
	}
}

func (c *Config) AddHost(m *data.Message) {
	debug.Ver("Config AddHost: %v", m)
	h := m.Data.(Host)

	// fire result event after function is done
	defer AfterCommand(m)

	if e := h.IsSane(c.Data, "subnet,port"); e != nil {
		// something is wrong with specified data
		m.Message = e.Error()
		m.Succeeded = false
	} else {
		for _, s := range c.Data.Servers {
			for _, n := range s.Networks {
				if n.Name == h.Network {
					n.Hosts = append(n.Hosts, h)
					m.Succeeded = true
					return
				}
			}
		}
		// uhm... no network with this name found
		m.Succeeded = false
		m.Message = err.New(NetworkNotFound, h.Network).Error()
	}
}
