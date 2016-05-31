package main

import (
	"dws/config"
	//"dws/lxc"
	//"gopkg.in/lxc/go-lxc.v2"
	"log"
)

func main() {
	dws_config.GatherConfig()
	log.Printf(dws_config.GetNetworks())
}
