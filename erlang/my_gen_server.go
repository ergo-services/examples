package main

import (
	"ergo.services/ergo/gen"
	"ergo.services/proto/erlang23"
)

func factory_GS() gen.ProcessBehavior {
	return &myGS{}
}

// to handle Erlang's cast messages and call request the erlnag23.GenServer
// must be used for the implementation.

type myGS struct {
	erlang23.GenServer
}

// HandleInfo invoked on messages sent from Erlang using erlang:send(...)
// or using gen.Process.Send(...)
func (m *myGS) HandleInfo(message any) error {
	m.Log().Info("got message: %v", message)
	return nil
}

// HandleCast invoked on messages sent from Erlang using gen_server:cast(...)
func (m *myGS) HandleCast(message any) error {
	m.Log().Info("got cast message: %v", message)
	return nil
}

// HandleCall invoked on request made using gen_server:call(...) or gen.Process.Call(...)
func (m *myGS) HandleCall(from gen.PID, ref gen.Ref, request any) (any, error) {
	m.Log().Info("got request from %s: %v", from, request)
	return m.PID(), nil
}
