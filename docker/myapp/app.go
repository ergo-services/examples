package myapp

import (
	"ergo.services/ergo/gen"
)

// CreateApp creates the myapp application behavior
func CreateApp() gen.ApplicationBehavior {
	return &MyApp{}
}

// MyApp implements gen.ApplicationBehavior
type MyApp struct{}

// Load is invoked on loading application
func (app *MyApp) Load(node gen.Node, args ...any) (gen.ApplicationSpec, error) {
	return gen.ApplicationSpec{
		Name:        "myapp",
		Description: "Example application with myactor",
		Mode:        gen.ApplicationModePermanent,
		Group: []gen.ApplicationMemberSpec{
			{
				Name:    "myactor",
				Factory: factory,
				Args:    []any{},
			},
		},
	}, nil
}

// Start is invoked once the application started
func (app *MyApp) Start(mode gen.ApplicationMode) {}

// Terminate is invoked once the application stopped
func (app *MyApp) Terminate(reason error) {}

// factory creates new instances of myActor
func factory() gen.ProcessBehavior {
	return &myActor{}
}
