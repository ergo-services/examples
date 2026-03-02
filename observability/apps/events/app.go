package events

import (
	"ergo.services/ergo/gen"
)

const (
	appName gen.Atom = "events_scenario"
)

func CreateApp() gen.ApplicationBehavior {
	return &app{}
}

type app struct{}

func (a *app) Load(node gen.Node, args ...any) (gen.ApplicationSpec, error) {
	return gen.ApplicationSpec{
		Name:        appName,
		Description: "Events scenario: publishers and subscribers for event metrics",
		Mode:        gen.ApplicationModeTemporary,
		Group: []gen.ApplicationMemberSpec{
			{
				Name:    "events_sup",
				Factory: factoryEventsSup,
			},
		},
	}, nil
}

func (a *app) Start(mode gen.ApplicationMode) {}
func (a *app) Terminate(reason error)         {}
