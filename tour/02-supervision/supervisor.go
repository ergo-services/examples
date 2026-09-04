package main

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

// Init declares what runs under this supervisor. Everything here is the
// default: One For One (restart only the child that failed) and the default
// restart policy, which is Transient - restart on an abnormal exit, leave a
// clean one alone. A panic is abnormal, so the worker comes back.
//
// The defaults also bound it: at most 5 restarts in any 5 seconds. Exceed that
// and the supervisor stops trying, which is the second half of this example.
func (s *supervisor) Init(args ...any) (act.SupervisorSpec, error) {
	spec := act.SupervisorSpec{
		Type: act.SupervisorTypeOneForOne,
		Children: []act.SupervisorChildSpec{
			{
				Name:    "worker",
				Factory: factoryWorker,
			},
		},
	}
	s.Log().Info("supervisor %s started", s.PID())
	return spec, nil
}
