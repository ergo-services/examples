package taskrunner

import (
	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

func factorySupervisor() gen.ProcessBehavior {
	return &supervisor{}
}

type supervisor struct {
	act.Supervisor
}

func (s *supervisor) Init(args ...any) (act.SupervisorSpec, error) {
	spec := act.SupervisorSpec{
		Type: act.SupervisorTypeOneForOne,
		Children: []act.SupervisorChildSpec{
			{Name: workerName, Factory: factoryWorker},
		},
	}

	s.Log().Info("supervisor %s started", s.PID())
	return spec, nil
}
