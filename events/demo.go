package main

import (
	"ergo.services/ergo"
	"ergo.services/ergo/gen"
	"ergo.services/logger/colored"
	"time"
)

func main() {
	var optionsPub, optionsSub gen.NodeOptions

	// disable default logger to get rid of multiple logging to the os.Stdout
	optionsPub.Log.DefaultLogger.Disable = true

	// add logger "colored".
	optionColored := colored.Options{TimeFormat: time.DateTime}
	loggerColored, err := colored.CreateLogger(optionColored)
	if err != nil {
		panic(err)
	}
	optionsPub.Log.Loggers = append(optionsPub.Log.Loggers, gen.Logger{Name: "cl", Logger: loggerColored})

	// set network cookie
	optionsPub.Network.Cookie = "123"

	// starting node for publishing
	nodePub, err := ergo.StartNode("node-pub@localhost", optionsPub)
	if err != nil {
		panic(err)
	}
	defer nodePub.StopForce()

	// spawn publisher process
	if _, err := nodePub.Spawn(factoryPub, gen.ProcessOptions{}); err != nil {
		panic(err)
	}

	// use the same cookie for the second node
	optionsSub.Network.Cookie = "123"
	// set the date/time format for the default logger
	optionsSub.Log.DefaultLogger.TimeFormat = time.DateTime
	// start second node
	nodeSub, err := ergo.StartNode("node-sub@localhost", optionsSub)
	if err != nil {
		panic(err)
	}

	// spawn 2 subscribers
	if _, err := nodeSub.Spawn(factorySub, gen.ProcessOptions{}); err != nil {
		panic(err)
	}
	// wait a bit before starting the second subscriber
	time.Sleep(time.Second)
	if _, err := nodeSub.Spawn(factorySub, gen.ProcessOptions{}); err != nil {
		panic(err)
	}

	nodeSub.Wait()
}
