package communication

import (
	"github.com/pfandl/dws/debug"
	"io"
	"net"
	"time"
)

type CanRead interface {
	Read(connection net.Conn) ([]byte, error)
}

type CanWrite interface {
	Write(connection net.Conn, data []byte) (int, error)
}

type CanListen interface {
	Listen(host string) error
}

type CanTalk interface {
	Talk(host string) error
}

type Thread struct {
	CanRead
	CanWrite
	CanListen
	CanTalk
	IsStoppable  bool
	HasInterface interface{}
	HasChannel   chan []byte
}

func (this *Thread) Listen(host string) error {
	debug.Ver("Thread Listen()")
	if l, err := net.Listen("tcp", ":"+host); err != nil {
		return err
	} else {
		// run in thread
		go this.Listener(l)
	}
	return nil
}

func (this *Thread) Talk(host string) error {
	debug.Ver("Thread Talk()")
	go this.Talker(host)
	return nil
}

func (this *Thread) Read(connection net.Conn) ([]byte, error) {
	var msg []byte

	// should be enough for most messages
	b := make([]byte, 2048)

	for {
		// we are looping to get all data
		s, err := connection.Read(b)
		if err != nil && err != io.EOF {
			debug.Warn("Thread CanRead read failed %s", err.Error())
			return nil, err
		} else {
			// EOF is a welcomed error
			if s <= 0 {
				// stop reading if EOF and no data left to read
				break
			}
			// add part of message
			msg = append(msg, b[0:s]...)
		}
	}

	return msg, nil
}

func (this *Thread) Listener(listener net.Listener) {
	debug.Ver("Thread Listener()")

	this.IsStoppable = true

	for this.IsStoppable {
		debug.Ver("Thread Listener Waiting...")

		connection, err := listener.Accept()
		if err != nil {
			debug.Err("Thread Listener connection failed %s", err.Error())
			continue
		}

		// connection established, further in an own thread
		go func(connection net.Conn) {
			debug.Ver("Thread Listener connection established with %s",
				connection.LocalAddr().String())

			if data, err := this.Read(connection); err != nil {
				debug.Err(err.Error())
			} else {
				this.HasChannel <- data
				result := <-this.HasChannel
				if size, err := this.Write(connection, result); err != nil {
					debug.Err(err.Error())
				} else {
					debug.Ver("Thread Listener wrote %d bytes - %v", size, result)
				}
			}

			// close when returning
			defer func() { connection.Close() }()
		}(connection)
	}
}

func (this *Thread) Talker(host string) {
	debug.Ver("Thread Talker(%s)", host)

	retries := 3

	this.IsStoppable = true
	for this.IsStoppable {
		// wait till we need to send data
		data := <-this.HasChannel

		var connection net.Conn
		var err error
		// try multiple times to establish a connection
		tried := 0
		for tried = 0; (tried == 0) || (err != nil) && (tried != retries); tried++ {
			// if not first attempt
			if tried > 0 {
				// let some time pass
				time.Sleep(1 * time.Second)
			}
			// dial
			connection, err = net.Dial("tcp", host)
		}

		if err != nil {
			debug.Err("Thread Talker connection failed %s", err.Error())
			this.HasChannel <- nil
			continue
		}

		debug.Ver("Thread Talker connection established with %s",
			connection.RemoteAddr().String())

		if size, err := this.Write(connection, data); err != nil {
			debug.Err(err.Error())
			this.HasChannel <- nil
		} else {
			debug.Ver("Thread Talker wroe %d bytes - %v", size, data)
		}
	}
}
