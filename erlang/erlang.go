package main

import (
	"flag"
	"fmt"
	"time"

	"ergo.services/application/observer"

	"ergo.services/logger/colored"

	"ergo.services/ergo"
	"ergo.services/ergo/gen"
	"ergo.services/proto/erlang23/dist"
	"ergo.services/proto/erlang23/epmd"
	"ergo.services/proto/erlang23/handshake"
)

var (
	OptionNodeName   string
	OptionNodeCookie string
)

func init() {
	flag.StringVar(&OptionNodeName, "name", "ergo@localhost", "node name")
	flag.StringVar(&OptionNodeCookie, "cookie", "123", "a secret cookie for the network messaging")
}

func main() {
	var options gen.NodeOptions

	flag.Parse()

	// add observer application
	apps := []gen.ApplicationBehavior{
		observer.CreateApp(observer.Options{}),
	}
	options.Applications = apps

	// disable default logger to get rid of multiple logging to the os.Stdout
	options.Log.DefaultLogger.Disable = true

	// add logger "colored"
	loggercolored, err := colored.CreateLogger(colored.Options{TimeFormat: time.DateTime})
	if err != nil {
		panic(err)
	}
	options.Log.Loggers = append(options.Log.Loggers, gen.Logger{Name: "colored", Logger: loggercolored})

	// set network cookie
	options.Network.Cookie = OptionNodeCookie

	// set the Erlang Network Stack for this node
	options.Network.Registrar = epmd.Create(epmd.Options{})
	options.Network.Handshake = handshake.Create(handshake.Options{})
	options.Network.Proto = dist.Create(dist.Options{})

	// starting node
	node, err := ergo.StartNode(gen.Atom(OptionNodeName), options)
	if err != nil {
		fmt.Printf("Unable to start node '%s': %s\n", OptionNodeName, err)
		return
	}

	// spawn process with the registered name.
	name := gen.Atom("myGenServer")
	pid, err := node.SpawnRegister(name, factory_GS, gen.ProcessOptions{})
	if err != nil {
		panic(err)
	}
	node.Log().Info("Spawned process %s with name %s based on erlang23.GenServer", pid, name)
	node.Log().Info("Now you can run Erlang node:")
	node.Log().Info("    $ erl -sname erl@localhost -setcookie %s", OptionNodeCookie)
	node.Log().Info("Or use docker image for that:")
	node.Log().Info("    $ docker run -it --network host --rm erlang:27-slim erl -sname erl@localhost -setcookie %s", OptionNodeCookie)
	node.Log().Info("To make a call request or send a message use the following commands:")
	node.Log().Info("    > erlang:send({%s, %s}, hello).", name, node.Name())
	node.Log().Info("    > gen_server:cast({%s, %s}, hi).", name, node.Name())
	node.Log().Info("    > gen_server:call({%s, %s}, request).", name, node.Name())

	node.Log().Info("Observer Application started and available at http://localhost:9911")
	node.Wait()
}
