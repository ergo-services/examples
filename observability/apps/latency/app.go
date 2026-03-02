package latency

import (
	"ergo.services/ergo/gen"
)

const (
	appName    gen.Atom = "latency_scenario"
	workerName gen.Atom = "latency_worker"
)

func CreateApp() gen.ApplicationBehavior {
	return &app{}
}

type app struct{}

func (a *app) Load(node gen.Node, args ...any) (gen.ApplicationSpec, error) {
	return gen.ApplicationSpec{
		Name:        appName,
		Description: "Latency scenario: periodic message bursts to remote worker pools",
		Mode:        gen.ApplicationModeTemporary,
		Map: map[string]gen.Atom{
			"worker": workerName,
		},
		Group: []gen.ApplicationMemberSpec{
			{
				Name:    "latency_sup",
				Factory: factoryLatencySup,
			},
		},
	}, nil
}

func (a *app) Start(mode gen.ApplicationMode) {}
func (a *app) Terminate(reason error)         {}
