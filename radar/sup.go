package demo

import (
	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

type workersSup struct {
	act.Supervisor
}

func factoryWorkersSup() gen.ProcessBehavior {
	return &workersSup{}
}

func (s *workersSup) Init(args ...any) (act.SupervisorSpec, error) {
	return act.SupervisorSpec{
		Type: act.SupervisorTypeOneForOne,
		Restart: act.SupervisorRestart{
			Strategy: act.SupervisorStrategyPermanent,
		},
		Children: []act.SupervisorChildSpec{
			{
				Name:    "db_worker",
				Factory: factoryDBWorker,
			},
			{
				Name:    "cache_worker",
				Factory: factoryCacheWorker,
			},
			{
				Name:    "api_worker",
				Factory: factoryAPIWorker,
			},
		},
	}, nil
}
