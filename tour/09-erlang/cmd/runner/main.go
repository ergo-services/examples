package main

import (
	"flag"
	"time"

	"ergo.services/ergo"
	"ergo.services/ergo/gen"
	"ergo.services/proto/erlang23/dist"
	"ergo.services/proto/erlang23/epmd"
	"ergo.services/proto/erlang23/handshake"

	"taskrunner/step09/apps/taskrunner"
)

var (
	name   = flag.String("name", "runner@localhost", "node name")
	cookie = flag.String("cookie", "taskrunner", "the Erlang cookie both nodes must share")
)

func main() {
	flag.Parse()

	var options gen.NodeOptions

	options.Log.DefaultLogger.TimeFormat = time.TimeOnly
	options.Network.Cookie = *cookie

	// This is the whole of it. Three components replace the native stack: EPMD
	// instead of the Ergo registrar, and the Erlang handshake and DIST protocol
	// instead of the Ergo ones. Nothing above this line knows.
	options.Network.Registrar = epmd.Create(epmd.Options{})
	options.Network.Handshake = handshake.Create(handshake.Options{})
	options.Network.Proto = dist.Create(dist.Options{})

	options.Applications = []gen.ApplicationBehavior{
		taskrunner.CreateApp(taskrunner.Options{}),
	}

	node, err := ergo.StartNode(gen.Atom(*name), options)
	if err != nil {
		panic(err)
	}

	if _, err := node.SpawnRegister(gateName, factoryGate, gen.ProcessOptions{}); err != nil {
		panic(err)
	}

	node.Log().Info("waiting for an Erlang node. Start one with:")
	node.Log().Info("    erl -sname erl@localhost -setcookie %s", *cookie)
	node.Log().Info("and submit a task with:")
	node.Log().Info("    gen_server:call({%s, %s}, {submit, <<\"task-1\">>, 150}).", gateName, node.Name())

	node.Wait()
}
