package main

import (
	"fmt"
	"os"
	"time"

	"ergo.services/ergo"
	"ergo.services/ergo/gen"
	"ergo.services/ergo/lib"
	"ergo.services/logger/colored"
	"ergo.services/registrar/etcd"

	"docker/myapp"
)

func main() {
	// Set node name from environment or default
	nodeName := os.Getenv("NODE_NAME")
	if nodeName == "" {
		nodeName = "node2@" + lib.GetHostname()
	}

	// Create etcd registrar
	registrarOptions := etcd.Options{
		Endpoints: []string{"etcd:2379"},
		Cluster:   "docker-example",
	}

	// Configure node options
	var options gen.NodeOptions
	registrar, err := etcd.Create(registrarOptions)
	if err != nil {
		fmt.Printf("Failed to create registrar: %v\n", err)
		return
	}
	options.Network.Registrar = registrar

	// Set network cookie to prevent warning about empty cookie
	options.Network.Cookie = "docker-example-cookie-123"

	// Create applications that must be started
	apps := []gen.ApplicationBehavior{
		myapp.CreateApp(),
	}
	options.Applications = apps

	// Disable default logger to get rid of multiple logging to the os.Stdout
	options.Log.DefaultLogger.Disable = true

	// Add logger "colored"
	loggercolored, err := colored.CreateLogger(colored.Options{TimeFormat: time.DateTime})
	if err != nil {
		fmt.Printf("Failed to create colored logger: %v\n", err)
		return
	}
	options.Log.Loggers = append(options.Log.Loggers, gen.Logger{Name: "colored", Logger: loggercolored})

	fmt.Printf("Starting %s...\n", nodeName)

	// Start the node
	node, err := ergo.StartNode(gen.Atom(nodeName), options)
	if err != nil {
		fmt.Printf("Failed to start node: %v\n", err)
		return
	}

	node.Log().Info("Node %s started successfully", nodeName)
	node.Log().Info("Application 'myapp' loaded and started automatically")
	node.Log().Info("Node2 is running...")

	// Wait for the node to stop
	node.Wait()
	node.Log().Info("Node2 stopped")
}
