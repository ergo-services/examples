package main

import (
	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

//
// Actor B
//

func factoryB() gen.ProcessBehavior {
	return &actorB{}
}

type actorB struct {
	act.Actor
}

func (b *actorB) Init(args ...any) error {
	b.Log().Info("started %s process on: %s", b.Name(), b.Node().Name())
	return nil
}

func (b *actorB) HandleCall(from gen.PID, ref gen.Ref, request any) (any, error) {
	switch r := request.(type) {
	case MyRequest:
		b.Log().Info("received MyRequest from %s: %#v", from, r)
		if r.MyBool {
			return 1, nil
		}
		return 2, nil
	}

	b.Log().Info("received unknown request: %#v", request)
	return false, nil
}
