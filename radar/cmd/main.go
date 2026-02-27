package main

import (
	"flag"
	"fmt"
	"time"

	demo "radar-demo"

	"ergo.services/application/radar"
	"ergo.services/ergo"
	"ergo.services/ergo/gen"
)

var (
	OptionNodeName string
)

func init() {
	flag.StringVar(&OptionNodeName, "name", "radar-demo@localhost", "node name")
}

func main() {
	flag.Parse()

	radarApp := radar.CreateApp(radar.Options{
		Port:                   9090,
		MetricsCollectInterval: 5 * time.Second,
		MetricsPoolSize:        3,
	})

	options := gen.NodeOptions{
		Applications: []gen.ApplicationBehavior{
			radarApp,
			demo.CreateWorkersApp(),
		},
	}

	node, err := ergo.StartNode(gen.Atom(OptionNodeName), options)
	if err != nil {
		fmt.Printf("unable to start node: %s\n", err)
		return
	}

	node.Log().Info("radar-demo started")
	node.Log().Info("  health:  http://localhost:9090/health/live")
	node.Log().Info("  health:  http://localhost:9090/health/ready")
	node.Log().Info("  health:  http://localhost:9090/health/startup")
	node.Log().Info("  metrics: http://localhost:9090/metrics")
	node.Log().Info("press Ctrl+C to stop")

	node.Wait()
}
