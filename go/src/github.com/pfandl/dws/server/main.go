package main

import (
	"encoding/json"
	"github.com/pfandl/dws"
	"io"
	"log"
	"net"
	"strings"
)

type Command struct {
	Name    string
	Execute func() string
}

var (
	Commands = []Command{
		{
			Name:    "get-networks",
			Execute: GetNetworks,
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

func server(ln net.Listener, ch chan string) {
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
		go server(ln, ch)
		// wait till client sends something
		msg := <-ch

		// check whether we now this command
		known := false
		for _, cmd := range Commands {
			if cmd.Name == msg {
				known = true
				// execute command
				res = cmd.Execute()
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
			// do nothing
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

func GetNetworks() string {
	if res, err := dws.GetNetworks(); err == nil {
		return Jsonify(res, nil)
	}
	return Jsonify(nil, dws.ErrConfigNetworkNoneAvailable)
}
