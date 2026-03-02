package latency

import (
	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

type latencySup struct {
	act.Supervisor
}

func factoryLatencySup() gen.ProcessBehavior {
	return &latencySup{}
}

func (s *latencySup) Init(args ...any) (act.SupervisorSpec, error) {
	return act.SupervisorSpec{
		Type: act.SupervisorTypeOneForOne,
		Restart: act.SupervisorRestart{
			Strategy: act.SupervisorStrategyPermanent,
		},
		Children: []act.SupervisorChildSpec{
			{
				Name:    workerName,
				Factory: factoryWorker,
			},
			{
				Name:    "latency_sender",
				Factory: factorySender,
			},
		},
	}, nil
}
