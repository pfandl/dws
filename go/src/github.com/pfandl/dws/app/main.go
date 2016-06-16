package main

import (
	"github.com/pfandl/dws/backingstore"
	"github.com/pfandl/dws/config"
	"github.com/pfandl/dws/data"
	"github.com/pfandl/dws/debug"
	"github.com/pfandl/dws/event"
	"github.com/pfandl/dws/module"
	"github.com/pfandl/dws/network"
	"github.com/pfandl/dws/server"
	"os"
	"os/signal"
)

func main() {
	debug.SetLevel(debug.All)
	debug.OnFatal(func() {
		module.StopAll()
	})
	module.Register(&config.Config{})
	module.Register(&backingstore.BackingStore{})
	module.Register(&server.Server{})
	module.Register(&network.Network{})
	if err := module.StartAll(); err != nil {
		debug.Fat(err.Error())
	}
	if err := module.GetError("config"); err != nil {
		debug.Fat(err.Error())
	}
	if err := module.GetError("backingstore"); err != nil {
		debug.Fat(err.Error())
	}
	if err := module.GetError("server"); err != nil {
		debug.Fat(err.Error())
	}
	if err := module.GetError("network"); err != nil {
		debug.Fat(err.Error())
	}
	event.Fire("command", &data.Message{Message: "get-backingstore-path"})

	// test adding server
	event.Fire(
		"command",
		&data.Message{
			Data: config.Server{
				Name: "test",
				IpV4: config.ServerIpV4{
					IpV4: config.IpV4{
						Port: "8002",
					},
				},
			},
			Message: "add-server",
		})
	/*
		// test adding network
		event.Fire(
			"command",
			&data.Message{
				Data: config.Network{
					Name: "test",
					IpV4: config.NetworkIpV4{
						IpV4: config.IpV4{
							Address: "1.1.1.1",
							Subnet:  "2.2.2.2",
						},
					},
					Server: "dws-test-server",
					Type:   "backup",
				},
				Message: "add-network",
			})

		// test adding host
		event.Fire(
			"command",
			&data.Message{
				Data: config.Host{
					Name: "test",
					IpV4: config.HostIpV4{
						IpV4: config.IpV4{
							Address: "129.168.1.1",
							Mac:     "aa:bb:cc:dd:ee:ff",
						},
					},
					UtsName: "bla.com",
					Network: "dws-production",
				},
				Message: "add-host",
			})
	*/
	// we are done loading, we now can just wait until we
	// get killed or gracefully stopped via system signals

	// set up signal interrupting, notifies on all signals
	c := make(chan os.Signal, 1)
	signal.Notify(c)

	// Block until a signal is received.
	_ = <-c

	debug.Info("signal received, stopping")

	module.StopAll()
}
