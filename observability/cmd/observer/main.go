package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"ergo.services/application/observer"
	"ergo.services/application/radar"
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

	// TEMPORARY: debug shows "session %s: inspect call %s took %s", which is how
	// the subscribe timeouts get measured. Drop this line afterwards.
	options.Log.Level = gen.LogLevelDebug
	// TEMPORARY: the framework inspectors hard-code Log().SetLogger("default"),
	// so disabling the default logger throws away every error they report.
	options.Log.DefaultLogger.Disable = false
	loggercolored, err := colored.CreateLogger(colored.Options{TimeFormat: time.DateTime})
	if err != nil {
		fmt.Printf("Failed to create colored logger: %v\n", err)
		return
	}
	options.Log.Loggers = append(options.Log.Loggers, gen.Logger{Name: "colored", Logger: loggercolored})

	// A listener rather than plain Host/Port: the surfaces and the ceiling are configured
	// there, and those are what a public deployment sets differently. Every variable below is
	// empty by default, which leaves the behaviour docker compose has always had.
	listener := observer.Listener{
		Host: "0.0.0.0",
		Port: 9911,
	}

	// the prefix a proxy in front passes through, "/observability" and the like
	listener.Path = os.Getenv("OBSERVER_PATH")

	// off when the interface comes from elsewhere - the cloud - and only the API is wanted
	if os.Getenv("OBSERVER_UI") == "off" {
		listener.UI.Disable = true
	}

	// origins allowed to call this observer, for a bundle served from another address
	if origins := os.Getenv("OBSERVER_ORIGINS"); origins != "" {
		listener.AllowedOrigins = strings.Split(origins, ",")
	}

	// how many SSE and MCP streams this observer holds at once, together. Default 64
	if streams := os.Getenv("OBSERVER_MAX_STREAMS"); streams != "" {
		n, err := strconv.Atoi(streams)
		if err != nil || n < 1 {
			fmt.Printf("Invalid OBSERVER_MAX_STREAMS %q\n", streams)
			os.Exit(1)
		}
		listener.MaxStreams = n
	}

	opts := observer.Options{Listeners: []observer.Listener{listener}}

	// On the whole observer, not the listener: a listener can be given less than the ceiling,
	// never more. Without it every visitor to a public demo can kill processes.
	if os.Getenv("OBSERVER_READONLY") == "1" {
		opts.Ceiling.ReadOnly = true
	}

	options.Applications = []gen.ApplicationBehavior{
		// the observer is a node of the cluster like any other, and Prometheus scrapes it
		// here: without radar the one process everybody is watching through is the one
		// nothing watches
		radar.CreateApp(radar.Options{Host: "0.0.0.0", Port: 9090}),
		observer.CreateApp(opts),
	}

	fmt.Printf("Starting %s...\n", nodeName)

	node, err := ergo.StartNode(gen.Atom(nodeName), options)
	if err != nil {
		fmt.Printf("Failed to start node: %v\n", err)
		return
	}

	node.Wait()
}
