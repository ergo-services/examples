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
		// Rest For One, not One For One as in step 02. The auditor is linked to
		// the worker's event, so it cannot outlive the worker and has to
		// subscribe again to whatever worker comes next. Declared after the
		// worker, it belongs to "the rest" and restarts with it. Nothing
		// depends on the auditor, so one that dies alone takes nothing along.
		Type: act.SupervisorTypeRestForOne,
		Children: []act.SupervisorChildSpec{
			{Name: workerName, Factory: factoryWorker},
			{Name: auditorName, Factory: factoryAuditor},
		},
	}

	s.Log().Info("supervisor %s started", s.PID())
	return spec, nil
}
