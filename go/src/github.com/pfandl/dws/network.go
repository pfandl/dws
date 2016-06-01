package dws

import (
	"github.com/milosgajdos83/tenus"
	"log"
	"net"
	"strconv"
	"strings"
)

func InitNetworking() error {
	v, err := GetNetworks()
	if err != nil {
		return err
	}
	networks, ok := v.([]Network)
	if ok == false {
		return ErrCannotConvertData
	}
	for _, n := range networks {
		b, err := tenus.BridgeFromName(n.Name)
		if err != nil {
			log.Printf("creating bridge %s", n.Name)
			b, err = tenus.NewBridgeWithName(n.Name)
			if err != nil {
				return err
			}
		}
		if err == nil {
			ip := net.ParseIP(n.IpV4.Address)
			if ip == nil {
				return ErrNetlinkCannotParseIpAddress
			}
			gw := net.ParseIP(n.Gateway.IpV4.Address)
			if gw == nil {
				return ErrNetlinkCannotParseGatewayIpAddress
			}
			subs := strings.Split(n.IpV4.Subnet, ".")
			aa, err := strconv.ParseInt(subs[0], 10, 64)
			bb, err := strconv.ParseInt(subs[1], 10, 64)
			cc, err := strconv.ParseInt(subs[2], 10, 64)
			dd, err := strconv.ParseInt(subs[3], 10, 64)
			var ipNet net.IPNet
			ipNet.IP = ip
			ipNet.Mask = net.IPv4Mask(
				byte(int32(aa)),
				byte(int32(bb)),
				byte(int32(cc)),
				byte(int32(dd)))
			log.Printf("configuring %s", n.Name)
			b.UnsetLinkIp(ip, &ipNet)
			/*
				log.Printf("setting gateway %s", gw.String())
				if err = b.SetLinkDefaultGw(&gw); err != nil {
					return err
				}
			*/
			log.Printf("setting ip %s", ipNet.String())
			if err = b.SetLinkIp(ip, &ipNet); err != nil {
				return err
			}
			log.Printf("upping %s", n.Name)
			if err = b.SetLinkUp(); err != nil {
				return err
			}
		}
	}
	return nil
}
