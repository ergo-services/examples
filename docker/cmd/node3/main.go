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
)

func main() {
	// Set node name from environment or default
	nodeName := os.Getenv("NODE_NAME")
	if nodeName == "" {
		nodeName = "node3@" + lib.GetHostname()
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

	// Create client application that spawns the client actor
	clientApp := &ClientApp{}
	apps := []gen.ApplicationBehavior{clientApp}
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
	node.Log().Info("Client application loaded with resolver actor")
	node.Log().Info("Node3 is running...")

	// Wait for the node to stop
	node.Wait()
	node.Log().Info("Node3 stopped")
}

// ClientApp implements gen.ApplicationBehavior for the client
type ClientApp struct{}

// Load is invoked on loading application
func (app *ClientApp) Load(node gen.Node, args ...any) (gen.ApplicationSpec, error) {
	return gen.ApplicationSpec{
		Name:        "client",
		Description: "Client application with resolver actor",
		Mode:        gen.ApplicationModeTemporary,
		Group: []gen.ApplicationMemberSpec{
			{
				Name:    "resolver",
				Factory: func() gen.ProcessBehavior { return &clientActor{} },
				Args:    []any{},
			},
		},
	}, nil
}

// Start is invoked once the application started
func (app *ClientApp) Start(mode gen.ApplicationMode) {}

// Terminate is invoked once the application stopped
func (app *ClientApp) Terminate(reason error) {}
