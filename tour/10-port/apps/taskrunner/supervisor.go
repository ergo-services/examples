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
		// Rest For One, and the order is the order of dependency. The workers
		// send to the auditor and the reporter subscribes to the event the
		// auditor owns, so the auditor comes first: if it dies, the two behind
		// it are restarted with it. The reporter is last and nothing depends on
		// it, so it can come and go alone - and it does, whenever the external
		// program it drives goes.
		Type: act.SupervisorTypeRestForOne,
		Children: []act.SupervisorChildSpec{
			{Name: auditorName, Factory: factoryAuditor, Args: args},
			{Name: workerName, Factory: factoryPool},
			{Name: reporterName, Factory: factoryReporter},
		},
	}

	s.Log().Info("supervisor %s started", s.PID())
	return spec, nil
}
