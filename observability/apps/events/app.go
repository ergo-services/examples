package events

import (
	"ergo.services/ergo/app"
	"ergo.services/ergo/gen"
)

const (
	appName gen.Atom = "events_scenario"
)

func CreateApp() gen.ApplicationBehavior {
	return &eventsApp{}
}

type eventsApp struct {
	app.Application
}

func (a *eventsApp) Load(args ...any) (gen.ApplicationSpec, error) {
	return gen.ApplicationSpec{
		Name:        appName,
		Description: "Events scenario: publishers and subscribers for event metrics",
		Mode:        gen.ApplicationModeTemporary,
		Network: gen.ApplicationNetwork{
			RegisterTypes: []any{MessageEventData{}},
		},
		Group: []gen.ApplicationMemberSpec{
			{
				Name:    "events_sup",
				Factory: factoryEventsSup,
			},
		},
	}, nil
}
