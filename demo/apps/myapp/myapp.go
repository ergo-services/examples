package myapp

import (
	"ergo.services/ergo/gen"
)

func CreateMyApp() gen.ApplicationBehavior {
	return &MyApp{}
}

type MyApp struct{}

// Load invoked on loading application using method ApplicationLoad of gen.Node interface.
func (app *MyApp) Load(node gen.Node, args ...any) (gen.ApplicationSpec, error) {
	return gen.ApplicationSpec{
		Name:        "myapp",
		Description: "description of this application",
		Mode:        gen.ApplicationModeTransient,
		Group: []gen.ApplicationMemberSpec{
			{
				Name:    "mysup",
				Factory: factory_MySup,
			},
		},
	}, nil
}

// Start invoked once the application started
func (app *MyApp) Start(mode gen.ApplicationMode) {}

// Terminate invoked once the application stopped
func (app *MyApp) Terminate(reason error) {}
