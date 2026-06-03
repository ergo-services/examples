package forest

import (
	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

type rootSupervisor struct {
	act.Supervisor
}

func factoryRoot() gen.ProcessBehavior {
	return &rootSupervisor{}
}

func (s *rootSupervisor) Init(args ...any) (act.SupervisorSpec, error) {
	return act.SupervisorSpec{
		Type: act.SupervisorTypeOneForOne,
		Restart: act.SupervisorRestart{
			Strategy:  act.SupervisorStrategyPermanent,
			Intensity: 10,
			Period:    5,
		},
		Children: []act.SupervisorChildSpec{
			{Name: computeSup, Factory: factoryComputeSup},
			{Name: ingestSup, Factory: factoryIngestSup},
			{Name: jobsSup, Factory: factoryJobsSup},
		},
	}, nil
}
