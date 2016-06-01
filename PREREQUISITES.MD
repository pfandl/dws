# Prerequisites for the DWS server #

## Centos 7 ##

	golang
	epel-release
	debootstrap
	perl
	libvirt
	lxc
	lxc-libs
	lxc-extra
	lxc-templates

## Fedora 23 ##

	golang
	debootstrap
	perl
	libvirt
	lxc
	lxc-libs
	lxc-extra
	lxc-templates

## General ##

### On the host ###

#### Enable port forwarding ####
	echo 1 > /proc/sys/net/ipv4/ip_forward
	echo "net.ipv4.ip_forward=1" >> /etc/sysctl.conf
	sysctl -p /etc/sysctl.conf

#### Enable masquerading ####
	iptables -t nat -A POSTROUTING -o external -j MASQUERADE
