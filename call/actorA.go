package main

import (
	"time"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

type doCallRemote struct{}
type doCallLocal struct{}

//
// Actor A
//

func factoryA() gen.ProcessBehavior {
	return &actorA{}
}

type actorA struct {
	act.Actor
}

func (a *actorA) Init(args ...any) error {
	a.Log().Info("started %s process on: %s", a.Name(), a.Node().Name())
	return nil
}

func (a *actorA) HandleMessage(from gen.PID, message any) error {
	switch message.(type) {
	case doCallLocal:
		local := gen.Atom("b")
		a.Log().Info("making request to local process %s", local)
		if result, err := a.Call(local, MyRequest{MyString: "abc"}); err == nil {
			a.Log().Info("received result from local process %s: %#v", local, result)
		} else {
			a.Log().Error("call local process failed: %s", err)
		}
		a.SendAfter(a.PID(), doCallRemote{}, time.Second)
		return nil

	case doCallRemote:
		remote := gen.ProcessID{Name: "b", Node: "nodeB@localhost"}
		a.Log().Info("making request to remote process %s", remote)
		if result, err := a.Call(remote, MyRequest{MyBool: true, MyString: "def"}); err == nil {
			a.Log().Info("received result from remote process %s: %#v", remote, result)
		} else {
			a.Log().Error("call remote process failed: %s", err)
		}

		a.SendAfter(a.PID(), doCallLocal{}, time.Second)
		return nil

	}

	a.Log().Error("unknown message %#v", message)
	return nil
}
