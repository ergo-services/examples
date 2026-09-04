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

// Init passes its own args straight through to the auditor. A supervisor that
// forwards what it was given is the usual way a dependency reaches a process
// several levels down.
func (s *supervisor) Init(args ...any) (act.SupervisorSpec, error) {
	spec := act.SupervisorSpec{
		// Rest For One. The workers send to the auditor, so the auditor starts
		// first and the pool is "the rest": an auditor that dies takes the pool
		// with it, and a pool that dies is restarted alone.
		Type: act.SupervisorTypeRestForOne,
		Children: []act.SupervisorChildSpec{
			{Name: auditorName, Factory: factoryAuditor, Args: args},
			{Name: workerName, Factory: factoryPool},
		},
	}

	s.Log().Info("supervisor %s started", s.PID())
	return spec, nil
}
