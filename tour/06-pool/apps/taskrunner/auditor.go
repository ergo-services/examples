package taskrunner

import (
	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

func factoryAuditor() gen.ProcessBehavior {
	return &auditor{}
}

// auditor owns the event now, and that is the pool's fault. An event name is
// registered by one process, and there are several workers - so the thing that
// publishes has to be the one process that outlives them. The subscribers on
// the other node cannot tell the difference: an event belongs to a node, not to
// a PID, and a worker restarting no longer costs anyone a subscription.
type auditor struct {
	act.Actor

	token gen.Ref
}

func (a *auditor) Init(args ...any) error {
	token, err := a.RegisterEvent(eventName, gen.EventOptions{
		Notify: true,
		Buffer: 10,
	})
	if err != nil {
		return err
	}
	a.token = token

	a.Log().Info("auditor %s is ready, publishing %s", a.PID(), eventName)
	return nil
}

func (a *auditor) HandleMessage(from gen.PID, message any) error {
	switch m := message.(type) {
	case messageDone:
		// One writer for the file, however many workers there are.
		a.Application().Behavior().(*taskrunnerApp).Record(m.ID, m.Took)

		return a.SendEvent(eventName, a.token, EventTaskDone{
			ID:   m.ID,
			For:  gen.Atom(m.For),
			Took: m.Took,
		})

	// Notify: true asked for these two. They fire when the number of
	// subscribers leaves zero and when it returns to zero. Nothing on this node
	// subscribes any more, so both are about the other node now.
	case gen.MessageEventStart:
		a.Log().Info("%s has listeners now", m.Name)

	case gen.MessageEventStop:
		a.Log().Info("%s has no listeners left", m.Name)
	}
	return nil
}

func (a *auditor) Terminate(reason error) {
	a.Log().Info("auditor %s stopped: %s", a.PID(), reason)
}
