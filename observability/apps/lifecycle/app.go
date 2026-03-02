package lifecycle

import (
	"ergo.services/ergo/gen"
)

const (
	appName   gen.Atom = "lifecycle_scenario"
	childSpec gen.Atom = "lifecycle_child"
)

func CreateApp() gen.ApplicationBehavior {
	return &app{}
}

type app struct{}

func (a *app) Load(node gen.Node, args ...any) (gen.ApplicationSpec, error) {
	return gen.ApplicationSpec{
		Name:        appName,
		Description: "Lifecycle scenario: spawn/terminate churn via SOFO supervisor",
		Mode:        gen.ApplicationModeTemporary,
		Group: []gen.ApplicationMemberSpec{
			{
				Name:    "lifecycle_sup",
				Factory: factoryLifecycleSup,
			},
			{
				Name:    "zombie_maker",
				Factory: factoryZombieMaker,
			},
		},
	}, nil
}

func (a *app) Start(mode gen.ApplicationMode) {}
func (a *app) Terminate(reason error)         {}
