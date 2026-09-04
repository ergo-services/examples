package main

import (
	"flag"
	"os"
	"time"

	"ergo.services/application/observer"
	"ergo.services/ergo"
	"ergo.services/ergo/gen"

	"taskrunner/step03/apps/taskrunner"
)

var failInit = flag.Bool("fail-init", false, "make the application fail in Init")

func main() {
	flag.Parse()

	var options gen.NodeOptions

	options.Log.DefaultLogger.TimeFormat = time.TimeOnly
	options.Network.Mode = gen.NetworkModeDisabled

	// Applications listed here are loaded and started by the node itself, in
	// order. The supervisor is no longer spawned by hand: it is a member of the
	// taskrunner group, so the application spawns it.
	options.Applications = []gen.ApplicationBehavior{
		observer.CreateApp(observer.Options{PoolSize: 4}),
		taskrunner.CreateApp(*failInit),
	}

	node, err := ergo.StartNode("tasks@localhost", options)
	if err != nil {
		// The node logged the reason and stopped itself. An application in
		// NodeOptions is not optional: if it cannot start, neither does the node.
		os.Exit(1)
	}

	if _, err := node.Spawn(factoryClient, gen.ProcessOptions{}); err != nil {
		panic(err)
	}

	node.Wait()
}
