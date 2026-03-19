package main

import (
	"fmt"
	"os"
	"time"

	"ergo.services/ergo"
	"ergo.services/ergo/gen"
	"ergo.services/ergo/lib"

	"ergo.services/application/mcp"
	"ergo.services/application/radar"
	"ergo.services/logger/colored"
	"ergo.services/registrar/etcd"

	"observability/apps/latency"
	"observability/apps/events"
	"observability/apps/lifecycle"
	"observability/apps/messaging"
)

func main() {
	nodeName := os.Getenv("NODE_NAME")
	if nodeName == "" {
		nodeName = fmt.Sprintf("node%d@%s", 3, lib.GetHostname())
	}

	registrar, err := etcd.Create(etcd.Options{
		Endpoints: []string{"etcd:2379"},
		Cluster:   "observability",
	})
	if err != nil {
		fmt.Printf("Failed to create registrar: %v\n", err)
		return
	}

	var options gen.NodeOptions
	options.Network.Registrar = registrar
	options.Network.Cookie = "observability-cookie"

	options.Log.Level = gen.LogLevelDebug
	options.Log.DefaultLogger.Disable = true
	loggercolored, err := colored.CreateLogger(colored.Options{TimeFormat: time.DateTime})
	if err != nil {
		fmt.Printf("Failed to create colored logger: %v\n", err)
		return
	}
	options.Log.Loggers = append(options.Log.Loggers, gen.Logger{Name: "colored", Logger: loggercolored})

	options.Applications = []gen.ApplicationBehavior{
		radar.CreateApp(radar.Options{Host: "0.0.0.0", Port: 9090}),
		mcp.CreateApp(mcp.Options{}),
		latency.CreateApp(),
		messaging.CreateApp(),
		lifecycle.CreateApp(),
		events.CreateApp(),
	}

	fmt.Printf("Starting %s...\n", nodeName)

	node, err := ergo.StartNode(gen.Atom(nodeName), options)
	if err != nil {
		fmt.Printf("Failed to start node: %v\n", err)
		return
	}

	node.Wait()
}
