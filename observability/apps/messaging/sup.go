package messaging

import (
	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

type messagingSup struct {
	act.Supervisor
}

func factoryMessagingSup() gen.ProcessBehavior {
	return &messagingSup{}
}

func (s *messagingSup) Init(args ...any) (act.SupervisorSpec, error) {
	return act.SupervisorSpec{
		Type: act.SupervisorTypeOneForOne,
		Restart: act.SupervisorRestart{
			Strategy: act.SupervisorStrategyPermanent,
		},
		Children: []act.SupervisorChildSpec{
			{
				Name:    poolName,
				Factory: factoryPool,
			},
			{
				Name:    "messaging_sender",
				Factory: factorySender,
			},
		},
	}, nil
}
