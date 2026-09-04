package main

import (
	"time"

	"ergo.services/application/observer"
	"ergo.services/ergo"
	"ergo.services/ergo/gen"

	"taskrunner/step06/apps/taskrunner"
)

func main() {
	var options gen.NodeOptions

	options.Log.DefaultLogger.TimeFormat = time.TimeOnly

	// The network is on now. Nothing else is needed to make this node
	// reachable: the name is the address, and the cookie is the shared secret
	// that decides who may connect.
	options.Network.Cookie = "taskrunner"

	options.Applications = []gen.ApplicationBehavior{
		observer.CreateApp(observer.Options{Port: 9911, PoolSize: 4}),
		taskrunner.CreateApp(),
	}

	node, err := ergo.StartNode("runner@localhost", options)
	if err != nil {
		panic(err)
	}

	node.Wait()
}
