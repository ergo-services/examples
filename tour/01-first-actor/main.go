package main

import (
	"time"

	"ergo.services/application/observer"
	"ergo.services/ergo"
	"ergo.services/ergo/gen"
)

func main() {
	var options gen.NodeOptions

	// Readable timestamps instead of the raw nanoseconds the default logger uses.
	options.Log.DefaultLogger.TimeFormat = time.TimeOnly

	// One node, no cluster yet: turn networking off so nothing else on this
	// machine can find us and the log stays about our own actors. Step 04 turns
	// it back on.
	options.Network.Mode = gen.NetworkModeDisabled

	// The observer gives every example a web UI. Open it while this runs: the
	// worker and the client are there, with their mailboxes. PoolSize is the
	// number of workers serving its HTTP requests - one browser tab does not
	// need the default 25, and a smaller pool keeps this log readable.
	options.Applications = []gen.ApplicationBehavior{
		observer.CreateApp(observer.Options{PoolSize: 4}),
	}

	node, err := ergo.StartNode("tasks@localhost", options)
	if err != nil {
		panic(err)
	}

	// A registered name makes the worker reachable as gen.Atom("worker") by
	// anyone on the node. Without it you would have to pass its PID around.
	if _, err := node.SpawnRegister("worker", factoryWorker, gen.ProcessOptions{}); err != nil {
		panic(err)
	}

	if _, err := node.Spawn(factoryClient, gen.ProcessOptions{}); err != nil {
		panic(err)
	}

	// Wait blocks until the node stops. Ctrl+C stops it.
	node.Wait()
}
