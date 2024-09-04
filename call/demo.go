package main

import (
	"ergo.services/ergo"
	"ergo.services/ergo/gen"
	"ergo.services/logger/colored"
	"time"
)

func main() {
	var optionsA, optionsB gen.NodeOptions

	// disable default logger to get rid of multiple logging to the os.Stdout
	optionsA.Log.DefaultLogger.Disable = true

	// add logger "colored".
	optionColored := colored.Options{TimeFormat: time.DateTime}
	loggerColored, err := colored.CreateLogger(optionColored)
	if err != nil {
		panic(err)
	}
	optionsA.Log.Loggers = append(optionsA.Log.Loggers, gen.Logger{Name: "cl", Logger: loggerColored})

	// set network cookie
	optionsA.Network.Cookie = "123"

	// starting nodeA
	nodeA, err := ergo.StartNode("nodeA@localhost", optionsA)
	if err != nil {
		panic(err)
	}

	// use the same cookie for the second node
	optionsB.Network.Cookie = "123"
	// set the date/time format for the default logger
	optionsB.Log.DefaultLogger.TimeFormat = time.DateTime
	// start second node
	nodeB, err := ergo.StartNode("nodeB@localhost", optionsB)
	if err != nil {
		panic(err)
	}
	defer nodeB.StopForce()

	// spawn process 'b' on the second node
	nodeB.SpawnRegister("b", factoryB, gen.ProcessOptions{})

	// spawn processes 'a' and 'b' on the first one
	nodeA.SpawnRegister("b", factoryB, gen.ProcessOptions{})
	nodeA.SpawnRegister("a", factoryA, gen.ProcessOptions{})

	// send message to the process 'a' on the first node
	nodeA.Send(gen.Atom("a"), doCallLocal{})

	nodeA.Wait()
}
