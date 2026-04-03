package tracing

import (
	"ergo.services/ergo/gen"
)

const (
	appName     gen.Atom = "tracing_scenario"
	relayName   gen.Atom = "trace_relay"
	sinkName    gen.Atom = "trace_sink"
	workerName  gen.Atom = "trace_worker"
)

func CreateApp() gen.ApplicationBehavior {
	return &app{}
}

type app struct{}

func (a *app) Load(node gen.Node, args ...any) (gen.ApplicationSpec, error) {
	return gen.ApplicationSpec{
		Name:        appName,
		Description: "Tracing scenario: distributed message chains across cluster",
		Mode:        gen.ApplicationModeTemporary,
		Group: []gen.ApplicationMemberSpec{
			{
				Name:    "tracing_sup",
				Factory: factorySup,
			},
		},
	}, nil
}

func (a *app) Start(mode gen.ApplicationMode) {}
func (a *app) Terminate(reason error)         {}
