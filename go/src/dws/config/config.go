package dws_config

import (
	"dws/global"
	//"bytes"
	"encoding/xml"
	//"io"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	//"strings"
)

const (
	Production = iota
	Backup     = iota
	Temporary  = iota
)

var (
	// Paths to gather config information ascending in importance
	Paths = []string{
		"./config",
		"/etc/dws/config",
		"/etc/default/dws",
	}
	Types = map[int]string{
		Production: "production",
		Backup:     "backup",
		Temporary:  "temporary",
	}
	IpV4RegExp = regexp.MustCompile("^(\\d{1,3}\\.){3}\\d{1,3}$")
	MacRegExp  = regexp.MustCompile("^([a-fA-F0-9]{2}:){5}[a-fA-F0-9]{2}$")
	UtsRegExp  = regexp.MustCompile("^(([a-zA-Z0-9\\-_])+\\.)*([a-zA-Z0-9\\-_])+\\.([a-zA-Z])+$")
	Settings   = Config{}
)

type Host struct {
	XMLName        xml.Name `xml:"host"`
	Name           string   `xml:"name,attr"`
	IpV4Address    string   `xml:"ipv4addr"`
	IpV4MacAddress string   `xml:"macaddr"`
	UtsName        string   `xml:"utsname"`
}

type Network struct {
	XMLName     xml.Name `xml:"network"`
	Name        string   `xml:"name,attr"`
	IpV4Gateway string   `xml:"gateway"`
	IpV4Subnet  string   `xml:"subnet"`
	Type        string   `xml:"type"`
	Hosts       []Host   `xml:"host"`
}

type Config struct {
	XMLName  xml.Name  `xml:"config"`
	Networks []Network `xml:"network"`
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
				Settings = conf
				break
			}
		}
	}
	return nil
}

func MergeConfig(c *Config) error {
	return nil
}

func IsSaneNetwork(n *Network, c *Config) error {
	if n.Name == "" {
		return dws_global.ErrConfigNetworkNameEmpty
	}
	if n.IpV4Gateway == "" {
		return dws_global.ErrConfigNetworkIpV4GatewayEmpty
	}
	if n.IpV4Subnet == "" {
		return dws_global.ErrConfigNetworkIpV4SubnetEmpty
	}
	if IpV4RegExp.MatchString(n.IpV4Gateway) == false {
		return dws_global.ErrConfigNetworkIpV4GatewaySyntax
	}
	if IpV4RegExp.MatchString(n.IpV4Subnet) == false {
		return dws_global.ErrConfigNetworkIpV4SubnetSyntax
	}
	if c != nil {
		for _, network := range c.Networks {
			if network.Name == n.Name {
				return dws_global.ErrConfigNetworkNameUsed
			}

		}
	}
	return nil
}

func IsSaneHost(h *Host, n *Network) error {
	if h.Name == "" {
		return dws_global.ErrConfigHostNameEmpty
	}
	if h.IpV4Address == "" {
		return dws_global.ErrConfigHostIpV4AddressEmpty
	}
	if h.IpV4MacAddress == "" {
		return dws_global.ErrConfigHostIpV4MacAddressEmpty
	}
	if h.UtsName == "" {
		return dws_global.ErrConfigHostUtsNameEmpty
	}
	if IpV4RegExp.MatchString(h.IpV4Address) == false {
		return dws_global.ErrConfigHostIpV4AddressSyntax
	}
	if MacRegExp.MatchString(h.IpV4MacAddress) == false {
		return dws_global.ErrConfigHostIpV4MacAddressSyntax
	}
	if UtsRegExp.MatchString(h.UtsName) == false {
		return dws_global.ErrConfigHostUtsNameSyntax
	}
	if n != nil {
		// do not allow same host name and same uts name on one network
		for _, host := range n.Hosts {
			if h.Name == host.Name {
				return dws_global.ErrConfigHostNameUsed
			}
			if h.UtsName == host.UtsName {
				return dws_global.ErrConfigHostUtsNameUsed
			}
		}
	}
	// do not allow same ip and same mac address across networks
	for _, network := range Settings.Networks {
		for _, host := range network.Hosts {
			if h.IpV4Address == host.IpV4Address {
				return dws_global.ErrConfigHostIpV4AddressUsed
			}
			if h.IpV4MacAddress == host.IpV4MacAddress {
				return dws_global.ErrConfigHostIpV4MacAddressUsed
			}
		}
	}
	return nil
}

func GetNetworks() (interface{}, error) {
	if len(Settings.Networks) == 0 {
		return nil, dws_global.ErrConfigNetworkNoneAvailable
	}
	return (Settings.Networks), nil
}

func GetHosts(n string) ([]string, error) {
	if n == "" {
		return nil, dws_global.ErrConfigHostNameNotFound
	}
	if len(Settings.Networks) == 0 {
		return nil, dws_global.ErrConfigNetworkNoneAvailable
	}
	var res []string
	for _, network := range Settings.Networks {
		res = append(res, network.Name)
	}
	return res, nil
}
