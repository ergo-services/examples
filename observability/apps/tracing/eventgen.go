package tracing

import (
	"fmt"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

// eventgen registers many events for testing the Observer events page pagination.
// Temporary: remove after testing.

type eventgen struct {
	act.Actor
}

func factoryEventGen() gen.ProcessBehavior {
	return &eventgen{}
}

func (eg *eventgen) Init(args ...any) error {
	for i := 0; i < 500; i++ {
		name := gen.Atom(fmt.Sprintf("test_event_%04d", i))
		eg.RegisterEvent(name, gen.EventOptions{})
	}
	eg.Log().Info("registered 500 test events")
	return nil
}

func (eg *eventgen) HandleMessage(from gen.PID, message any) error {
	return nil
}

func (eg *eventgen) Terminate(reason error) {}
