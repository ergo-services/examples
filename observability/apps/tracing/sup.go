package tracing

import (
	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

type sup struct {
	act.Supervisor
}

func factorySup() gen.ProcessBehavior {
	return &sup{}
}

func (s *sup) Init(args ...any) (act.SupervisorSpec, error) {
	return act.SupervisorSpec{
		Type: act.SupervisorTypeOneForOne,
		Restart: act.SupervisorRestart{
			Strategy: act.SupervisorStrategyPermanent,
		},
		Children: []act.SupervisorChildSpec{
			{
				// relay: receives calls, forwards to sink on another node
				Name:    relayName,
				Factory: factoryRelay,
			},
			{
				// sink: receives messages and calls from other nodes
				Name:    sinkName,
				Factory: factorySink,
			},
			{
				// worker: periodically sends messages, calls, and forward chains
				Name:    workerName,
				Factory: factoryWorker,
			},
			{
				// eventgen: registers many events for testing Observer pagination
				Name:    "event_gen",
				Factory: factoryEventGen,
			},
		},
	}, nil
}
