package main

import (
	//"fmt"
	"github.com/pfandl/dws"
	//"dws/lxc"
	//"gopkg.in/lxc/go-lxc.v2"
	"encoding/json"
	//"github.com/milosgajdos83/tenus"
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
	if err := dws.GatherConfig(); err != nil {
		log.Fatalf("aborting: %s", err.Error())
	}
	if err := dws.InitNetworking(); err != nil {
		log.Fatalf("aborting: %s", err.Error())
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
	if res, err := dws.GetNetworks(); err == nil {
		return Jsonify(res, nil)
	}
	return Jsonify(nil, dws.ErrConfigNetworkNoneAvailable)
}
