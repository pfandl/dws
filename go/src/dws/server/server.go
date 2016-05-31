package main

import (
	"dws/lxc"
	"gopkg.in/lxc/go-lxc.v2"
	"log"
)

var (
	lxc_config_path string
)

func main() {
	lxc_config_path := lxc.DefaultConfigPath()
	if lxc_config_path == "" {
		log.Fatalf("cannot obtain lxc default config path, is lxc installed?")
	}

	log.Printf("lxc config path: %s", lxc_config_path)

	res, err := dws_lxc.GetTemplates()
	if err != nil {
		log.Printf("Error: %s", err.Error())
	}
	log.Printf("templates: %s", res)
}
