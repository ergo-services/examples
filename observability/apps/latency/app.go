package latency

import (
	"ergo.services/ergo/app"
	"ergo.services/ergo/gen"
)

const (
	appName    gen.Atom = "latency_scenario"
	workerName gen.Atom = "latency_worker"
)

func CreateApp() gen.ApplicationBehavior {
	return &latencyApp{}
}

type latencyApp struct {
	app.Application
}

func (a *latencyApp) Load(args ...any) (gen.ApplicationSpec, error) {
	return gen.ApplicationSpec{
		Name:        appName,
		Description: "Latency scenario: periodic message bursts to remote worker pools",
		Mode:        gen.ApplicationModeTemporary,
		Map: map[string]gen.Atom{
			"worker": workerName,
		},
		Network: gen.ApplicationNetwork{
			RegisterTypes: []any{MessagePing{}},
		},
		Group: []gen.ApplicationMemberSpec{
			{
				Name:    "latency_sup",
				Factory: factoryLatencySup,
			},
		},
	}, nil
}
