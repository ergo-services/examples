package main

import (
	"fmt"
	"os"
	"time"

	"ergo.services/ergo"
	"ergo.services/ergo/gen"
	"ergo.services/ergo/lib"

	"ergo.services/application/mcp"
	"ergo.services/application/pulse"
	"ergo.services/application/radar"
	"ergo.services/logger/colored"
	"ergo.services/registrar/etcd"

	"observability/apps/events"
	"observability/apps/latency"
	"observability/apps/lifecycle"
	"observability/apps/messaging"
	"observability/apps/tracing"
)

func main() {
	nodeName := os.Getenv("NODE_NAME")
	if nodeName == "" {
		nodeName = fmt.Sprintf("node%d@%s", 1, lib.GetHostname())
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
		mcp.CreateApp(mcp.Options{Host: "0.0.0.0", Port: 9922}),
		pulse.CreateApp(pulse.Options{
			Endpoint: "tempo:4318",
			Insecure: true,
		}),
		latency.CreateApp(),
		messaging.CreateApp(),
		lifecycle.CreateApp(),
		events.CreateApp(),
		tracing.CreateApp(),
	}

	fmt.Printf("Starting %s...\n", nodeName)

	node, err := ergo.StartNode(gen.Atom(nodeName), options)
	if err != nil {
		fmt.Printf("Failed to start node: %v\n", err)
		return
	}

	node.Wait()
}
