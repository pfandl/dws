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
		"add-server-result",
		"remove-server-result",
		"add-backingstore-result",
		"remove-backingstore-result",
		"add-network-result",
		"remove-network-result",
		"add-host-result",
		"remove-host-result",
		// events fired after executing commands
		// when we are ok and other modules return the result to server
		"check-server-added",
		"check-server-removed",
		"check-backingstore-added",
		"check-backingstore-removed",
		"check-network-added",
		"check-network-removed",
		"check-host-added",
		"check-host-removed",
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
	// for ipv4 comparison
	// set to false if comparing hosts
	// set to true if comparing networks
	ComparingNetworks = true
	// our config data
	LoadedConfig *Config
	// errors
	NoConfig               = "no valid configuration found"
	ServerNameAlreadyUsed  = "server name is already used"
	NetworkNameAlreadyUsed = "network name is already used"
	InvalidNetworkType     = "network type is invalid"
	IpV4Overlap            = "network ip/subnet is overlapping with existing network"
	IpV4AlreadyUsed        = "ip is already in use"
	HostNameAlreadyUsed    = "host name is already in use"
	UtsNameAlreadyUsed     = "uts name is already in use"
	WrongSubnet            = "subnets do not match"
	ServerNotFound         = "server not found"
	NetworkNotFound        = "network not found"
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
	Address string   `xml:"address"   validation:"ipv4"`
	Subnet  string   `xml:"subnet"    validation:"ipv4"`
	Mac     string   `xml:"mac"       validation:"ipv4mac"`
	Port    string   `xml:"port"      validation:"port"`
}

type Host struct {
	Propagate
	SaneConfig
	XMLName xml.Name `xml:"host"`
	Name    string   `xml:"name,attr" validation:"!empty"`
	IpV4    IpV4     `xml:"ipv4"      validation:"struct"         validation-ignore:"subnet,port"`
	UtsName string   `xml:"utsname"   validation:"uts"`
	Network string
}

type Network struct {
	Propagate
	SaneConfig
	XMLName xml.Name `xml:"network"`
	Name    string   `xml:"name,attr" validation:"!empty,max=15"`
	IpV4    IpV4     `xml:"ipv4"      validation:"struct"         validation-ignore:"mac,port"`
	Type    string   `xml:"type"      validation:"!empty"`
	Hosts   []Host   `xml:"host"      validation:"slice"`
	Server  string
}

type Log struct {
	Propagate
	SaneConfig
	XMLName xml.Name `xml:"log"`
}

type Server struct {
	Propagate
	SaneConfig
	XMLName  xml.Name  `xml:"server"`
	Name     string    `xml:"name,attr" validation:"!empty"`
	IpV4     IpV4      `xml:"ipv4"      validation:"struct"         validation-ignore:"address,subnet,mac"`
	Networks []Network `xml:"network"   validation:"slice"`
	Log      Log       `xml:"log"`
}

type ConfigData struct {
	Propagate
	SaneConfig
	XMLName      xml.Name `xml:"config"`
	Name         string   `xml:"name,attr" validation:"!empty"`
	Servers      []Server `xml:"server" validation:"slice"`
	ValidateData bool
}

func (d *ConfigData) Available() error {
	debug.Ver("ConfigData: Available")
	for i := 0; i < len(d.Servers); i++ {
		if err := d.Servers[i].Available(); err != nil {
			return err
		}
	}
	return nil
}

func (d *ConfigData) IsSane(c *ConfigData, s string) error {
	debug.Ver("ConfigData: IsSane")
	for i := 0; i < len(c.Servers); i++ {
		srv := &c.Servers[i]
		if err := srv.IsSane(c, s); err != nil {
			return err
		}
	}

	// need to validate data thats not read out from the config file
	if c.ValidateData == true {
		if err := validation.Validate(*d, s, ""); err != nil {
			return err
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
	return nil
}

func (d *Server) IsSane(c *ConfigData, s string) error {
	debug.Ver("Server: IsSane")

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

	// need to validate data thats not read out from the config file
	if c.ValidateData == true {
		if err := validation.Validate(*d, s, ""); err != nil {
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
	if Networks[d.Type] == 0 {
		return err.New(InvalidNetworkType, d.Type)
	}

	// check ip

	// set comparison method for ipv4 IsSane method
	ComparingNetworks = true
	// but also reset it after function returns
	defer func() { ComparingNetworks = false }()

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

	// need to validate data thats not read out from the config file
	if c.ValidateData == true {
		if err := validation.Validate(*d, s, ""); err != nil {
			return err
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

	// need to validate data thats not read out from the config file
	if c.ValidateData == true {
		if err := validation.Validate(*d, s, ""); err != nil {
			return err
		}
	}

	return nil
}

func (d *IpV4) IsSane(c *ConfigData, s string) error {
	debug.Ver("IpV4: IsSane")
	for i := 0; i < len(c.Servers); i++ {
		srv := &c.Servers[i]
		for i := 0; i < len(srv.Networks); i++ {
			n := &srv.Networks[i]
			// skip same object (do not compare to itself)
			if d == &n.IpV4 {
				continue
			}
			if ComparingNetworks {
				debug.Ver("IpV4: IsSane checking network ip %s %s", d.Address, n.IpV4.Address)
				if d.Address == n.IpV4.Address {
					return err.New(IpV4AlreadyUsed, d.Address)
				}
				debug.Ver("IpV4: IsSane checking network subnet %s", d.Subnet)
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

			} else {
				for i := 0; i < len(n.Hosts); i++ {
					h := &n.Hosts[i]
					// skip same object (do not compare to itself)
					if d == &h.IpV4 {
						continue
					}
					debug.Ver("IpV4: IsSane checking host ip %s %s", d.Address, h.IpV4.Address)
					if d.Address == h.IpV4.Address {
						return err.New(IpV4AlreadyUsed, d.Address)
					}
				}
			}
		}
	}

	// need to validate data thats not read out from the config file
	if c.ValidateData == true {
		if err := validation.Validate(*d, s, ""); err != nil {
			return err
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
		conf := &ConfigData{ValidateData: false}

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
		// start validating data in the IsSane functions
		c.Data.ValidateData = true

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

func (c *Config) Event(e string, v interface{}) {
	debug.Ver("Config got event: %s %v", e, v)
	switch e {
	case "add-server":
		c.AddServer(v.(*data.Message))
	case "add-network":
		c.AddNetwork(v.(*data.Message))
	case "add-host":
		c.AddHost(v.(*data.Message))
	default:
		debug.Fat("Config event %s unknown", e)
	}
}

func (c *Config) AddServer(m *data.Message) {
	debug.Ver("Config AddServer: %v", m)
	s := m.Data.(Server)

	// fire result event after function is done
	defer func() {
		if m.Succeeded == true {
			// host module should also check data
			event.Fire("check-server-added", m)
		} else {
			// something went wrong, inform server about that
			event.Fire("add-server-result", m)
		}
	}()

	if e := s.IsSane(c.Data, "address,subnet,mac"); e != nil {
		// something is wrong with specified data
		m.Message = e.Error()
		m.Succeeded = false
	} else {
		c.Data.Servers = append(c.Data.Servers, s)
		m.Message = ServerAdded
		m.Succeeded = true
	}
}

func (c *Config) AddNetwork(m *data.Message) {
	debug.Ver("Config AddNetwork: %v", m)
	n := m.Data.(Network)

	// fire result event after function is done
	defer func() {
		if m.Succeeded == true {
			// host module should also check data
			event.Fire("check-network-added", m)
		} else {
			// something went wrong, inform server about that
			event.Fire("add-network-result", m)
		}
	}()

	if e := n.IsSane(c.Data, "mac,port"); e != nil {
		// something is wrong with specified data
		m.Message = e.Error()
		m.Succeeded = false
	} else {
		for _, s := range c.Data.Servers {
			if s.Name == n.Server {
				s.Networks = append(s.Networks, n)
				m.Message = NetworkAdded
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
	defer func() {
		if m.Succeeded == true {
			// host module should also check data
			event.Fire("check-host-added", m)
		} else {
			// something went wrong, inform server about that
			event.Fire("add-host-result", m)
		}
	}()

	if e := h.IsSane(c.Data, "subnet,port"); e != nil {
		// something is wrong with specified data
		m.Message = e.Error()
		m.Succeeded = false
	} else {
		for _, s := range c.Data.Servers {
			for _, n := range s.Networks {
				if n.Name == h.Network {
					n.Hosts = append(n.Hosts, h)
					m.Message = HostAdded
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
