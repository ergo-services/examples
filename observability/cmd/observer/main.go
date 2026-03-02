package main

import (
	"fmt"
	"os"
	"time"

	"ergo.services/application/observer"
	"ergo.services/ergo"
	"ergo.services/ergo/gen"
	"ergo.services/logger/colored"
	"ergo.services/registrar/etcd"
)

func main() {
	nodeName := os.Getenv("NODE_NAME")
	if nodeName == "" {
		nodeName = "observer@observer"
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

	options.Log.DefaultLogger.Disable = true
	loggercolored, err := colored.CreateLogger(colored.Options{TimeFormat: time.DateTime})
	if err != nil {
		fmt.Printf("Failed to create colored logger: %v\n", err)
		return
	}
	options.Log.Loggers = append(options.Log.Loggers, gen.Logger{Name: "colored", Logger: loggercolored})

	options.Applications = []gen.ApplicationBehavior{
		observer.CreateApp(observer.Options{
			Host: "0.0.0.0",
			Port: 9911,
		}),
	}

	fmt.Printf("Starting %s...\n", nodeName)

	node, err := ergo.StartNode(gen.Atom(nodeName), options)
	if err != nil {
		fmt.Printf("Failed to start node: %v\n", err)
		return
	}

	node.Wait()
}
