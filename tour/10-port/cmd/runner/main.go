package main

import (
	"time"

	"ergo.services/application/observer"
	"ergo.services/ergo"
	"ergo.services/ergo/gen"

	"taskrunner/step10/apps/taskrunner"
)

// what an agent is told about this node before it asks anything
const instructions = `This node runs the taskrunner application from the Ergo
tour. Tasks arrive at the registered process taskrunner_worker, which is a pool
of workers; taskrunner_auditor records every finished task and publishes the
'tasks' event. Tasks come from client@localhost.`

func main() {
	var options gen.NodeOptions

	options.Log.DefaultLogger.TimeFormat = time.TimeOnly
	options.Network.Cookie = "taskrunner"

	options.Applications = []gen.ApplicationBehavior{
		observer.CreateApp(observer.Options{
			PoolSize: 4,

			// One listener, spelled out rather than left to the Port default,
			// because the MCP surface on it is narrowed.
			Listeners: []observer.Listener{
				{
					Port: 9911,
					MCP: observer.SurfaceMCP{
						// An agent gets to look, not to touch. The browser on
						// the same listener still has everything.
						Ceiling:      &observer.Ceiling{ReadOnly: true},
						Instructions: instructions,
					},
				},
			},
		}),
		taskrunner.CreateApp(taskrunner.Options{}),
	}

	node, err := ergo.StartNode("runner@localhost", options)
	if err != nil {
		panic(err)
	}

	node.Wait()
}
