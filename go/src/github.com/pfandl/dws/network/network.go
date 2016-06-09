package network

import (
	"github.com/milosgajdos83/tenus"
	"github.com/pfandl/dws/config"
	"github.com/pfandl/dws/data"
	"github.com/pfandl/dws/debug"
	"github.com/pfandl/dws/error"
	"github.com/pfandl/dws/event"
	"github.com/pfandl/dws/module"
	"net"
	"os"
	"strconv"
	"strings"
)

var (
	// events we fire
	ActiveEvents = []string{
		"add-network-result",
		"remove-network-result",
	}
	// events we are interested in
	PassiveEvents = []string{
		"network-available",
		"check-network-added",
		"check-network-removed",
	}
	// errors
	CannotParseIpAddress = "cannot parse ip address"
	IpRemovedFromBridge  = "ip address was removed from bridge"
	SubnetMismatch       = "configured and actual bridge subnet mismatch"
	// messages
	NetworkAdded = "network was added"

	FixNetwork = false
)

type Network struct {
	module.Module
	Networks []*config.Network
}

func (c *Network) Name() string {
	return "network"
}

func (c *Network) Events(active bool) []string {
	debug.Ver("Network: Events %v", active)
	if active == true {
		return ActiveEvents
	} else {
		return PassiveEvents
	}
}

func (c *Network) Init() error {
	debug.Ver("Network Init()")
	// check arguments if fix-network was passed
	for _, a := range os.Args {
		if a == "fix-network" {
			FixNetwork = true
			break
		}
	}
	return nil
}

func (c *Network) Start() error {
	debug.Ver("Network Start()")
	// check all networks for existance
	for _, n := range c.Networks {
		debug.Ver("Network check existance %s", n.Name)
		// find bridge
		if b, e := tenus.BridgeFromName(n.Name); e != nil {
			debug.Warn("Network bridge not available %s %s", n.Name, e.Error())
			// fail if not overriden by arguments
			if FixNetwork == true {
				debug.Info("Network trying to fix")
				if e = c.CreateBridge(n); e != nil {
					return e
				}
			}
			return e
		} else {
			f := false
			var e error
			// find set ip address
			debug.Ver("Network check ip %s and subnet %s", n.IpV4.Address, n.IpV4.Subnet)
			a, e := b.NetInterface().Addrs()
			for _, i := range a {
				debug.Ver("Network has ip %s %s", i.String(), i.Network())
				// the ip we get is probably in xxx.xxx.xxx.xxx/xx notation
				if strings.Split(i.String(), "/")[0] == n.IpV4.Address {
					f = true
					break
				}
			}
			// try to fix network if wanted and necessary
			if FixNetwork == true {
				if e != nil || f == false {
					debug.Info("Network trying to fix")
					e = c.SetBridgeIp(&b, n)
					if e == nil {
						f = true
					}
				}
			} else {
				return e
			}
			// net error, none found or subnet mismatch
			if e != nil {
				return e
			}
			// config ok?
			if f == false {
				return err.New(IpRemovedFromBridge, n.IpV4.Address, n.Name)
			}
		}
	}
	return nil
}

func (c *Network) SetBridgeIp(b *tenus.Bridger, n *config.Network) error {
	debug.Ver("Network SetBridgeIp %v", n)
	// remove any ip
	// TODO
	/*
		if a, e := b.NetInterface().Addrs(); e == nil {
			for _, i := range a {
				net.Par
				if e = (*b).UnsetLinkIp(i); e != nil {
					return nil
				}
			}
		}*/
	// ip parsing
	ip := net.ParseIP(n.IpV4.Address)
	if ip == nil {
		return err.New(CannotParseIpAddress, n.IpV4.Address)
	}
	// subnet parsing
	subs := strings.Split(n.IpV4.Subnet, ".")
	aa, e := strconv.ParseInt(subs[0], 10, 64)
	bb, e := strconv.ParseInt(subs[1], 10, 64)
	cc, e := strconv.ParseInt(subs[2], 10, 64)
	dd, e := strconv.ParseInt(subs[3], 10, 64)
	var ipNet net.IPNet
	ipNet.IP = ip
	ipNet.Mask = net.IPv4Mask(
		byte(int32(aa)),
		byte(int32(bb)),
		byte(int32(cc)),
		byte(int32(dd)))
	debug.Ver("Network set ip address %v", ip)
	// set ip
	if e = (*b).SetLinkIp(ip, &ipNet); e != nil {
		return e
	}
	return nil
}

func (c *Network) CreateBridge(n *config.Network) error {
	debug.Ver("Network CreateBridge %v", n)
	b, e := tenus.NewBridgeWithName(n.Name)
	if e != nil {
		return e
	}
	debug.Ver("Network setting ip %s %s", n.IpV4.Address, n.IpV4.Subnet)
	if e = c.SetBridgeIp(&b, n); e != nil {
		return e
	}
	debug.Ver("Network upping link %v", b)
	if e = b.SetLinkUp(); e != nil {
		return e
	}
	return nil
}

func (c *Network) Stop() error {
	debug.Ver("Network Stop()")
	return nil
}

func (c *Network) Event(e string, v interface{}) {
	debug.Ver("Network got event: %s %v", e, v)
	switch e {
	case "network-available":
		c.Available(v.(*config.Network))
	case "check-network-added":
		c.Added(v.(*data.Message))
	default:
		debug.Fat("Network event %s unknown", e)
	}
}

func (c *Network) Add(n *config.Network) {
	debug.Ver("Network Add: %v", n)
	c.Networks = append(c.Networks, n)
}

func (c *Network) Added(m *data.Message) {
	debug.Ver("Network network available: %v", m.Data)
	n := m.Data.(config.Network)

	// fire result event after function is done
	defer func() {
		event.Fire("add-network-result", m)
	}()

	if err := c.CreateBridge(&n); err != nil {
		m.Succeeded = false
		m.Message = err.Error()
	} else {
		m.Succeeded = true
		m.Message = NetworkAdded
		c.Add(&n)
	}
}

func (c *Network) Available(n *config.Network) {
	debug.Ver("Network network available: %v", n)
	c.Add(n)
}
