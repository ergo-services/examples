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
		// Still Rest For One, but the dependency turned around. The workers
		// send to the auditor, so the auditor starts first and the pool is
		// "the rest": an auditor that dies takes the pool with it, and a pool
		// that dies is restarted alone.
		Type: act.SupervisorTypeRestForOne,
		Children: []act.SupervisorChildSpec{
			{Name: auditorName, Factory: factoryAuditor},
			{Name: workerName, Factory: factoryPool},
		},
	}

	s.Log().Info("supervisor %s started", s.PID())
	return spec, nil
}
