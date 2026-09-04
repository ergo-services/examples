package main

import (
	"time"

	"ergo.services/ergo/gen"
	"ergo.services/proto/erlang23"
	"ergo.services/proto/erlang23/etf"

	"taskrunner/step09/apps/taskrunner"
)

const gateName gen.Atom = "taskrunner_gate"

func factoryGate() gen.ProcessBehavior {
	return &gate{}
}

// gate is the whole of the Erlang side. It embeds erlang23.GenServer, so
// Erlang's gen_server:call and gen_server:cast reach it as callbacks, and it
// turns what arrives into the ordinary package API from step 03.
//
// Nothing else in the tour changes. The boundary is one process, not the
// application: the pool, the auditor and the event have no idea.
type gate struct {
	erlang23.GenServer
}

// HandleCall answers gen_server:call. Erlang sends terms, so this is where a
// term becomes a Go value: {submit, Id, EffortMs} with Id as a binary and
// EffortMs as an integer.
func (g *gate) HandleCall(from gen.PID, ref gen.Ref, request any) (any, error) {
	tuple, ok := request.(etf.Tuple)
	if ok == false || len(tuple) != 3 || tuple.Element(1) != gen.Atom("submit") {
		g.Log().Warning("cannot read %#v", request)
		return etf.Tuple{gen.Atom("error"), gen.Atom("badarg")}, nil
	}

	// The types are the ones ETF produces, not the ones a Go author would
	// guess: an Erlang binary arrives as []byte and an integer as int64. Only
	// atoms come through as their Ergo counterpart, gen.Atom.
	id, ok := tuple.Element(2).([]byte)
	if ok == false {
		g.Log().Warning("id arrived as %T, want a binary", tuple.Element(2))
		return etf.Tuple{gen.Atom("error"), gen.Atom("badarg")}, nil
	}
	ms, ok := tuple.Element(3).(int64)
	if ok == false {
		g.Log().Warning("effort arrived as %T, want an integer", tuple.Element(3))
		return etf.Tuple{gen.Atom("error"), gen.Atom("badarg")}, nil
	}

	if err := taskrunner.Submit(g, g.Node().Name(), string(id), time.Duration(ms)*time.Millisecond); err != nil {
		g.Log().Error("cannot submit %s: %s", id, err)
		return etf.Tuple{gen.Atom("error"), gen.Atom("unavailable")}, nil
	}

	g.Log().Info("submitted %s for %s", id, from.Node)
	return gen.Atom("ok"), nil
}

// HandleCast answers gen_server:cast, which expects no reply.
func (g *gate) HandleCast(message any) error {
	g.Log().Info("cast: %#v", message)
	return nil
}

// HandleInfo answers a plain erlang:send.
func (g *gate) HandleInfo(message any) error {
	g.Log().Info("info: %#v", message)
	return nil
}
