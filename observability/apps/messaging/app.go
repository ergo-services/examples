package messaging

import (
	"ergo.services/ergo/gen"
)

const (
	appName  gen.Atom = "messaging_scenario"
	poolName gen.Atom = "messaging_pool"
)

func CreateApp() gen.ApplicationBehavior {
	return &app{}
}

type app struct{}

func (a *app) Load(node gen.Node, args ...any) (gen.ApplicationSpec, error) {
	return gen.ApplicationSpec{
		Name:        appName,
		Description: "Messaging scenario: random bursts with variable payload size",
		Mode:        gen.ApplicationModeTemporary,
		Map: map[string]gen.Atom{
			"worker": poolName,
		},
		Group: []gen.ApplicationMemberSpec{
			{
				Name:    "messaging_sup",
				Factory: factoryMessagingSup,
			},
		},
	}, nil
}

func (a *app) Start(mode gen.ApplicationMode) {}
func (a *app) Terminate(reason error)         {}
