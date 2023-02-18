package main

import (
	"fmt"
	"time"

	"github.com/ergo-services/ergo/etf"
	"github.com/ergo-services/ergo/gen"
)

type MyPing struct {
	gen.Server
}

type messageInfo struct{}
type messageCast struct{}
type messageCall struct{}

func (p *MyPing) Init(process *gen.ServerProcess, args ...etf.Term) error {
	process.CastAfter(process.Self(), messageInfo{}, time.Second)
	return nil
}

func (p *MyPing) HandleCast(process *gen.ServerProcess, message etf.Term) gen.ServerStatus {
	switch message.(type) {
	case messageInfo:
		fmt.Println("MyPing send message 'Hello World'")
		process.Send(poolProcessName, "Hello World")

		// schedule sending cast message
		process.CastAfter(process.Self(), messageCast{}, time.Second)
	case messageCast:
		fmt.Println("MyPing cast message 'Hello World'")
		process.Cast(poolProcessName, "Hello World")

		// schedule making a call request
		process.CastAfter(process.Self(), messageCall{}, time.Second)
	case messageCall:
		fmt.Println("MyPing make call request 'ping'")
		result, err := process.Call(poolProcessName, "ping")
		if err != nil {
			panic(err)
		}
		if result != "pong" {
			panic("wrong result")
		}

		// schedule sending a regular message
		process.CastAfter(process.Self(), messageInfo{}, time.Second)
	}

	return gen.ServerStatusOK
}
