package dws

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"regexp"
	"strconv"
)

const (
	ProductionNetwork = iota
	BackupNetwork     = iota
	TemporaryNetwork  = iota
)

const (
	BackingStoreBtrfs = iota
)

var (
	// Paths to gather config information ascending in importance
	Paths = []string{
		"./config",
		"/etc/dws/config",
		"/etc/default/dws",
	}
	NetworkTypes = map[int]string{
		ProductionNetwork: "production",
		BackupNetwork:     "backup",
		TemporaryNetwork:  "temporary",
	}
	BackingStoreTypes = map[int]string{
		BackingStoreBtrfs: "btrfs",
	}
	IpV4RegExp = regexp.MustCompile("^(\\d{1,3}\\.){3}\\d{1,3}$")
	MacRegExp  = regexp.MustCompile("^([a-fA-F0-9]{2}:){5}[a-fA-F0-9]{2}$")
	UtsRegExp  = regexp.MustCompile("^(([a-zA-Z0-9\\-_])+\\.)*([a-zA-Z0-9\\-_])+\\.([a-zA-Z])+$")
	Settings   = Config{}
)

type IpV4 struct {
	XMLName xml.Name `xml:"ipv4"`
	Address string   `xml:"address"`
	Subnet  string   `xml:"subnet"`
	Mac     string   `xml:"mac"`
	Port    string   `xml:"port"`
}

type Host struct {
	XMLName xml.Name `xml:"host"`
	Name    string   `xml:"name,attr"`
	IpV4    IpV4     `xml:"ipv4"`
	UtsName string   `xml:"utsname"`
}

type Gateway struct {
	XMLName xml.Name `xml:"gateway"`
	IpV4    IpV4     `xml:"ipv4"`
}

type Network struct {
	XMLName xml.Name `xml:"network"`
	Name    string   `xml:"name,attr"`
	Gateway Gateway  `xml:"gateway"`
	IpV4    IpV4     `xml:"ipv4"`
	Type    string   `xml:"type"`
	Hosts   []Host   `xml:"host"`
}

type TcpHost struct {
	XMLName xml.Name `xml:"host"`
	IpV4    IpV4     `xml:"ipv4"`
}

type BackingStore struct {
	XMLName xml.Name `xml:"backingstore"`
	Host    TcpHost  `xml:"host"`
	Path    string   `xml:"path"`
	Type    string   `xml:"type"`
}

type Server struct {
	XMLName xml.Name `xml:"server"`
	Host    TcpHost  `xml:"host"`
}

type Config struct {
	XMLName      xml.Name     `xml:"config"`
	BackingStore BackingStore `xml:"backingstore"`
	Server       Server       `xml:"server"`
	Networks     []Network    `xml:"network"`
}

func WriteConfig() {
	enc := xml.NewEncoder(os.Stdout)
	enc.Indent("  ", "    ")
	fmt.Printf(xml.Header)
	if err := enc.Encode(&Settings); err != nil {
		fmt.Printf("error: %v\n", err)
	}
	fmt.Printf("\n</xml>")
}

func GatherConfig() error {
	for _, path := range Paths {
		log.Printf("reading config from %s", path)
		data, err := ioutil.ReadFile(path)
		if err != nil {
			log.Printf("could not read config file from %s (%v)", path, err)
		} else {
			conf := Config{}
			err = xml.Unmarshal(data, &conf)
			if err != nil {
				log.Printf("could not parse config file %s (%v)", path, err)
			} else {
				if err = IsSaneConfig(&conf); err == nil {
					log.Printf("using config file from %s", path)
					Settings = conf
					return nil
				} else {
					log.Printf(err.Error())
				}
			}
		}
	}
	return ErrConfigInvalid
}

func IsSaneConfig(c *Config) error {
	if c == nil {
		c = &Settings
	}
	// check backing store
	if c.BackingStore.Host.IpV4.Address == "" &&
		c.BackingStore.Host.IpV4.Port != "" {
		// we are the backing store, check for sane type
		if c.BackingStore.Type == "" {
			return ErrConfigBackingStoreTypeEmpty
		}
		var t string
		for _, _t := range BackingStoreTypes {
			if _t == c.BackingStore.Type {
				t = _t
				break
			}
		}
		if t == "" {
			return ErrConfigBackingStoreTypeInvalid
		}
		i, err := strconv.ParseInt(c.BackingStore.Host.IpV4.Port, 10, 64)
		if err != nil || i > math.MaxUint16 {
			return ErrConfigBackingStorePortInvalid
		}
	} else {
		// check server
		if c.Server.Host.IpV4.Port == "" {
			return ErrConfigServerPortEmpty
		} else {
			i, err := strconv.ParseInt(c.Server.Host.IpV4.Port, 10, 64)
			if err != nil || i > math.MaxUint16 {
				return ErrConfigServerPortInvalid
			}
		}

	}
	// check networks (implies hosts)
	for i := 0; i < len(c.Networks); i++ {
		n := &c.Networks[i]
		if err := IsSaneNetwork(n, c); err != nil {
			return err
		}
	}
	return nil
}

func IsSaneNetwork(n *Network, c *Config) error {
	if n.Name == "" {
		return ErrConfigNetworkNameEmpty
	}
	log.Printf("checking network %s", n.Name)
	if n.Gateway.IpV4.Address == "" {
		return ErrConfigNetworkIpV4GatewayEmpty
	}
	if n.IpV4.Address == "" {
		return ErrConfigNetworkIpV4Empty
	}
	if n.IpV4.Subnet == "" {
		return ErrConfigNetworkIpV4SubnetEmpty
	}
	if IpV4RegExp.MatchString(n.Gateway.IpV4.Address) == false {
		return ErrConfigNetworkIpV4GatewaySyntax
	}
	if IpV4RegExp.MatchString(n.IpV4.Address) == false {
		return ErrConfigNetworkIpV4Syntax
	}
	if IpV4RegExp.MatchString(n.IpV4.Subnet) == false {
		return ErrConfigNetworkIpV4SubnetSyntax
	}
	if c != nil {
		for i := 0; i < len(c.Networks); i++ {
			network := &c.Networks[i]
			// skip checking against self
			if network == n {
				continue
			}
			if network.Name == n.Name {
				return ErrConfigNetworkNameUsed
			}
			for i := 0; i < len(network.Hosts); i++ {
				host := &network.Hosts[i]
				if err := IsSaneHost(host, network); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func IsSaneHost(h *Host, n *Network) error {
	if h.Name == "" {
		return ErrConfigHostNameEmpty
	}
	log.Printf("checking host %s", h.Name)
	if h.IpV4.Address == "" {
		return ErrConfigHostIpV4AddressEmpty
	}
	if h.IpV4.Mac == "" {
		return ErrConfigHostIpV4MacAddressEmpty
	}
	if h.UtsName == "" {
		return ErrConfigHostUtsNameEmpty
	}
	if IpV4RegExp.MatchString(h.IpV4.Address) == false {
		return ErrConfigHostIpV4AddressSyntax
	}
	if MacRegExp.MatchString(h.IpV4.Mac) == false {
		return ErrConfigHostIpV4MacAddressSyntax
	}
	if UtsRegExp.MatchString(h.UtsName) == false {
		return ErrConfigHostUtsNameSyntax
	}
	if n != nil {
		// do not allow same host name and same uts name on one network
		for i := 0; i < len(n.Hosts); i++ {
			host := &n.Hosts[i]
			// skip checking against self
			if host == h {
				continue
			}
			if h.Name == host.Name {
				return ErrConfigHostNameUsed
			}
			if h.UtsName == host.UtsName {
				return ErrConfigHostUtsNameUsed
			}
		}
	}
	// do not allow same ip and same mac address across networks
	for _, network := range Settings.Networks {
		for _, host := range network.Hosts {
			if h.IpV4.Address == host.IpV4.Address {
				return ErrConfigHostIpV4AddressUsed
			}
			if h.IpV4.Mac == host.IpV4.Mac {
				return ErrConfigHostIpV4MacAddressUsed
			}
		}
	}
	return nil
}

func GetNetworks() (interface{}, error) {
	if len(Settings.Networks) == 0 {
		return nil, ErrConfigNetworkNoneAvailable
	}
	return (Settings.Networks), nil
}

func GetHosts(n string) ([]string, error) {
	if n == "" {
		return nil, ErrConfigHostNameNotFound
	}
	if len(Settings.Networks) == 0 {
		return nil, ErrConfigNetworkNoneAvailable
	}
	var res []string
	for _, network := range Settings.Networks {
		res = append(res, network.Name)
	}
	return res, nil
}

func IsBackingStoreHost() bool {
	return Settings.BackingStore.Host.IpV4.Address == "" &&
		Settings.BackingStore.Host.IpV4.Port != ""
}

func HasBackingStore() bool {
	return Settings.BackingStore.Host.IpV4.Address != "" &&
		Settings.BackingStore.Host.IpV4.Port != ""
}

func GetConnection() string {
	return Settings.Server.Host.IpV4.Address +
		":" +
		Settings.Server.Host.IpV4.Port
}

func GetBackingStoreConnection() string {
	return Settings.BackingStore.Host.IpV4.Address +
		":" +
		Settings.BackingStore.Host.IpV4.Port
}
