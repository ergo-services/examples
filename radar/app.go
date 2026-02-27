package demo

import (
	"ergo.services/application/radar"
	"ergo.services/ergo/gen"
)

func CreateWorkersApp() gen.ApplicationBehavior {
	return &workersApp{}
}

type workersApp struct{}

func (a *workersApp) Load(node gen.Node, args ...any) (gen.ApplicationSpec, error) {
	return gen.ApplicationSpec{
		Name:        "workers",
		Description: "demo workers that use Radar health and metrics",
		Mode:        gen.ApplicationModeTransient,
		Depends: gen.ApplicationDepends{
			Applications: []gen.Atom{radar.Name},
		},
		Group: []gen.ApplicationMemberSpec{
			{
				Name:    "workers_sup",
				Factory: factoryWorkersSup,
			},
		},
	}, nil
}

func (a *workersApp) Start(mode gen.ApplicationMode) {}
func (a *workersApp) Terminate(reason error)         {}
