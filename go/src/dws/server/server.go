package main

import (
	"dws/config"
	"dws/global"
	"fmt"
	//"dws/lxc"
	//"gopkg.in/lxc/go-lxc.v2"
	"encoding/json"
	"github.com/milosgajdos83/tenus"
	"log"
)

type Command struct {
	Name    string
	Execute func() interface{}
}

var (
	Commands = []Command{
		{
			Name:    "get-networks",
			Execute: GetNetworks,
		},
	}
)

func main() {
	dws_config.GatherConfig()
	log.Printf("%s", GetNetworks())

	// Create a new network bridge
	br, err := tenus.NewBridgeWithName("mybridge")
	if err != nil {
		log.Fatal(err)
	}

	// Bring the bridge up
	if err = br.SetLinkUp(); err != nil {
		fmt.Println(err)
	}

	// Create a dummy link
	dl, err := tenus.NewLink("mydummylink")
	if err != nil {
		log.Fatal(err)
	}

	// Add the dummy link into bridge
	if err = br.AddSlaveIfc(dl.NetInterface()); err != nil {
		log.Fatal(err)
	}

	// Bring the dummy link up
	if err = dl.SetLinkUp(); err != nil {
		fmt.Println(err)
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
	if data, err := json.Marshal(v); err == nil {
		return string(data)
	}
	return `{Succeeded: false, Message: "cannot convert to JSON"}`
}

func GetNetworks() interface{} {
	if res, err := dws_config.GetNetworks(); err == nil {
		return Jsonify(res, nil)
	}
	return Jsonify(nil, dws_global.ErrCannotConvertData)
}
