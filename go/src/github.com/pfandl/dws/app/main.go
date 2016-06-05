package main

import (
	"encoding/json"
	"github.com/pfandl/dws"
	"github.com/pfandl/dws/config"
	"github.com/pfandl/dws/debug"
	"github.com/pfandl/dws/module"
	"github.com/pfandl/dws/network"
	"github.com/pfandl/dws/server"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	//"time"
)

type Command struct {
	Name    string
	Execute func([]string) string
}

var (
	Commands = []Command{
		{
			Name:    "get-networks",
			Execute: GetNetworks,
		},
		{
			Name:    "add-network",
			Execute: AddNetwork,
		},
		{
			Name:    "init-backingstore",
			Execute: InitBackingStore,
		},
	}
	BackingStoreChannel = make(chan string)
)

func client(cn net.Conn, ch chan string) {
	var msg string

	// we dont need a big buffer i guess
	buf := make([]byte, 256)

	for {
		// wait till server sends data
		msg = <-ch
		b := []byte(msg)
		s := len(b)

		// writing
		for {
			// we are looping until we write all data
			i, err := cn.Write([]byte(msg))
			if err != nil {
				log.Printf("write failed %s", err.Error())
				break
			} else {
				if i <= 0 {
					// stop writing if no error and no bytes written
					break
				}
				// decrease amount to write
				s -= i
			}
			// is everything written yet?
			if s == 0 {
				break
			}
		}

		// reading
		for {
			// we are looping to get all data
			s, err := cn.Read(b)
			if err != nil && err != io.EOF {
				log.Printf("read failed %s", err.Error())
				break
			} else {
				// EOF is a welcomed error
				if s <= 0 {
					// stop reading if EOF and no data left to read
					break
				}
				// add part of message
				msg += string(buf[0:s])
			}
		}
		// did we receive anything?
		if msg != "" {
			// write to chanel
			ch <- strings.Replace(msg, "\n", "", -1)
			// reset message for next communication
			msg = ""
		}
		// and go back waiting for new data to be sent
	}
}

func _server(ln net.Listener, ch chan string) {
	var msg string

	// we dont need a big buffer i guess
	b := make([]byte, 256)

	for {
		// forever, accept connections
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("connection failed %s", err.Error())
		}
		for {
			// we are looping to get all data
			s, err := conn.Read(b)
			if err != nil && err != io.EOF {
				log.Printf("read failed %s", err.Error())
				break
			} else {
				// EOF is a welcomed error
				if s <= 0 {
					// stop reading if EOF and no data left to read
					break
				}
				// add part of message
				msg += string(b[0:s])
			}
		}
		// did we receive anything?
		if msg != "" {
			// write to chanel
			ch <- strings.Replace(msg, "\n", "", -1)
			// reset message for next communication
			msg = ""
		}
		// wait till response channel is filled and write data to client
		if _, err = conn.Write([]byte(<-ch)); err != nil {
			log.Printf("write failed %s", err.Error())
		}
		// done, close connection
		conn.Close()
		// and go back waiting for new connections
	}
}

func listenerThread(ln net.Listener) {
	// create a read write channel
	ch := make(chan string)
	var res string
	for {
		// start as thread
		go _server(ln, ch)
		// wait till client sends something
		msg := <-ch
		// split into arguments
		arg := strings.Split(msg, " ")

		// check whether we now this command
		known := false
		for _, cmd := range Commands {
			// actual command is first argument
			if cmd.Name == arg[0] {
				known = true
				// execute command
				res = cmd.Execute(arg[1:len(arg)])
			}
		}
		if known == false {
			res = "unknown command"
		}
		// write back to the client
		ch <- res
		res = ""
	}
}

func main() {
	debug.SetLevel(debug.All)
	module.Register(&config.Config{})
	module.Register(&server.Server{})
	module.Register(&network.Network{})
	if err := module.InitAll(); err != nil {
		debug.Fat(err.Error())
	}
	if err := module.GetError("config"); err != nil {
		debug.Fat(err.Error())
	}
	if err := module.GetError("server"); err != nil {
		debug.Fat(err.Error())
	}

	// we are done loading, we now can just wait until we
	// get killed or gracefully stopped via system signals

	// set up signal interrupting, notifies on all signals
	c := make(chan os.Signal, 1)
	signal.Notify(c)

	// Block until a signal is received.
	_ = <-c

	/*
		if err := dws.GatherConfig(); err != nil {
			log.Fatalf("aborting: %s", err.Error())
		}
		if dws.IsBackingStoreHost() == false {
			// we are the main host
			if err := dws.InitNetworking(); err != nil {
				log.Fatalf("aborting: %s", err.Error())
			}

			c := dws.GetConnection()
			log.Printf("starting server on %s", c)
			ln, err := net.Listen("tcp", c)
			if err != nil {
				log.Fatalf("aborting: ", err)
			}

			go listenerThread(ln)

			// connect to backing store host if necessary
			if dws.HasBackingStore() {
				c := dws.GetBackingStoreConnection()
				log.Printf("connecting to backing store on %s", c)
				ln, err := net.Dial("tcp", c)
				if err != nil {
					log.Fatalf("aborting: ", err)
				}
				go client(ln, BackingStoreChannel)
			}

			for {
				// sleep to not eat the cpu
				time.Sleep(1 * time.Second)
			}

		} else {
			// backing store host only has to provide
			// space for our lxc virtual machines
			c := dws.GetBackingStoreConnection()
			log.Printf("starting backing store on %s", c)
			ln, err := net.Listen("tcp", c)
			if err != nil {
				log.Fatalf("aborting: ", err)
			}

			listenerThread(ln)
		}
	*/
}

func Jsonify(v interface{}, err error) string {
	type Result struct {
		Succeeded bool
		Result    interface{}
		Message   string
	}
	res := Result{}
	if err != nil {
		res.Succeeded = false
		res.Message = err.Error()
	} else {
		res.Succeeded = true
		res.Result = v
	}
	if data, err := json.Marshal(res); err == nil {
		return string(data)
	}
	return `{Succeeded: false, Message: "cannot convert to JSON"}`
}

func GetNetworks([]string) string {
	if res, err := dws.GetNetworks(); err == nil {
		return Jsonify(res, nil)
	}
	return Jsonify(nil, dws.ErrConfigNetworkNoneAvailable)
}

func AddNetwork(args []string) string {
	if len(args) != 4 {
		return Jsonify(nil, dws.ErrInvalidParameter)
	}

	n := &dws.Network{}
	n.Name = args[0]
	n.IpV4.Address = args[1]
	n.IpV4.Subnet = args[2]
	n.Type = args[3]

	if err := dws.IsSaneNetwork(n, &dws.Settings); err != nil {
		return Jsonify(nil, err)
	}

	return Jsonify(nil, nil)
}

func InitBackingStore(args []string) string {
	if dws.IsBackingStoreHost() == true {
		if dws.IsBackingStoreBtrfs() {
			//dws.InitBtrfsBackingStore()
		}
	}
	return Jsonify(nil, nil)
}
