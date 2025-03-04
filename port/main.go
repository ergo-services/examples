package main

import (
	"flag"
	"fmt"
	"time"

	"ergo.services/logger/colored"
	"ergo.services/logger/rotate"

	"ergo.services/ergo"
	"ergo.services/ergo/gen"
	"ergo.services/ergo/lib"
)

var (
	OptionNodeName   string
	OptionNodeCookie string
	OptionModeBinary bool
	OptionModeTxt    bool
)

func init() {
	flag.StringVar(&OptionNodeName, "name", "demo@localhost", "node name")
	flag.StringVar(&OptionNodeCookie, "cookie", lib.RandomString(16), "a secret cookie for the network messaging")
	flag.BoolVar(&OptionModeBinary, "bin", false, "demo with binary mode (with auto-chunking)")
	flag.BoolVar(&OptionModeTxt, "txt", true, "demo with text mode")
}

func main() {
	var options gen.NodeOptions

	flag.Parse()

	// disable default logger to get rid of multiple logging to the os.Stdout
	options.Log.DefaultLogger.Disable = true

	// add logger "colored".
	coloredOptions := colored.Options{
		TimeFormat:  time.DateTime,
		IncludeName: true,
	}
	loggercolored, err := colored.CreateLogger(coloredOptions)
	if err != nil {
		panic(err)
	}
	options.Log.Loggers = append(options.Log.Loggers, gen.Logger{Name: "colored", Logger: loggercolored})

	// add logger "rotate".
	loggerrotate, err := rotate.CreateLogger(rotate.Options{TimeFormat: time.DateTime})
	if err != nil {
		panic(err)
	}
	options.Log.Loggers = append(options.Log.Loggers, gen.Logger{Name: "rotate", Logger: loggerrotate})

	// set network options
	options.Network.Cookie = OptionNodeCookie

	// starting node
	node, err := ergo.StartNode(gen.Atom(OptionNodeName), options)
	if err != nil {
		fmt.Printf("Unable to start node '%s': %s\n", OptionNodeName, err)
		return
	}

	// start txt actor
	if OptionModeTxt {
		if _, err := node.SpawnRegister("actor-txt", factory_ActorPortTxt, gen.ProcessOptions{}); err != nil {
			panic(err)
		}
	}

	// start bin actor
	if OptionModeBinary {
		if _, err := node.SpawnRegister("actor-bin", factory_ActorPortBin, gen.ProcessOptions{}); err != nil {
			panic(err)
		}
	}

	// wait node
	node.Wait()
}
