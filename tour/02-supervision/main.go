package main

import (
	"time"

	"ergo.services/application/observer"
	"ergo.services/ergo"
	"ergo.services/ergo/gen"
)

func main() {
	var options gen.NodeOptions

	options.Log.DefaultLogger.TimeFormat = time.TimeOnly
	options.Network.Mode = gen.NetworkModeDisabled
	options.Applications = []gen.ApplicationBehavior{
		observer.CreateApp(observer.Options{PoolSize: 4}),
	}

	node, err := ergo.StartNode("tasks@localhost", options)
	if err != nil {
		panic(err)
	}

	// The worker is no longer spawned here. The supervisor owns it: it spawns
	// it, registers the name, and puts it back when it dies.
	if _, err := node.Spawn(factorySupervisor, gen.ProcessOptions{}); err != nil {
		panic(err)
	}

	if _, err := node.Spawn(factoryClient, gen.ProcessOptions{}); err != nil {
		panic(err)
	}

	node.Wait()
}
