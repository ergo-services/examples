package taskrunner

import (
	"errors"

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

func (s *supervisor) HandleMessage(from gen.PID, message any) error {
	switch message.(type) {
	case messageBreakSupervisor:
		// Returning an error from a callback terminates the process with that
		// error as the reason. This one is a member of the group, so the
		// application hears about it.
		return errors.New("simulated supervisor failure")
	}
	return nil
}
