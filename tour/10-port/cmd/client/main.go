package main

import (
	"flag"
	"time"

	"ergo.services/application/observer"
	"ergo.services/ergo"
	"ergo.services/ergo/gen"

	"taskrunner/step10/apps/taskrunner"
)

var runner = flag.String("runner", "runner@localhost", "node the task runner is on")

func main() {
	flag.Parse()

	var options gen.NodeOptions

	options.Log.DefaultLogger.TimeFormat = time.TimeOnly

	// Same cookie as the other node, or the handshake is refused.
	options.Network.Cookie = "taskrunner"

	options.Applications = []gen.ApplicationBehavior{
		observer.CreateApp(observer.Options{Port: 9912, PoolSize: 4}),
	}

	node, err := ergo.StartNode("client@localhost", options)
	if err != nil {
		panic(err)
	}

	// This node does not host the application, so nothing registered the wire
	// types for it. Without this the first Send fails to encode.
	if err := taskrunner.RegisterTypes(node); err != nil {
		panic(err)
	}

	if _, err := node.Spawn(factoryClient, gen.ProcessOptions{}, gen.Atom(*runner)); err != nil {
		panic(err)
	}

	node.Wait()
}
