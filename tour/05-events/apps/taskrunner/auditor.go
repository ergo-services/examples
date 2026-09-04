package taskrunner

import (
	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

func factoryAuditor() gen.ProcessBehavior {
	return &auditor{}
}

// auditor is the only process that writes the results file. It learns what
// happened from the event, not from the worker, so adding a second worker later
// changes nothing here.
type auditor struct {
	act.Actor
}

func (a *auditor) Init(args ...any) error {
	// LinkEvent, not MonitorEvent: this process has no reason to exist without
	// the event. If the worker goes, so does the auditor, and the supervisor
	// puts the pair back together.
	//
	// Subscribing returns whatever the producer's buffer already holds.
	// Throwing that away is how you lose the events that happened before you
	// arrived.
	buffered, err := a.LinkEvent(Tasks(a.Node().Name()))
	if err != nil {
		return err
	}
	for _, event := range buffered {
		a.record(event)
	}

	a.Log().Info("auditor %s is ready, %d event(s) were waiting", a.PID(), len(buffered))
	return nil
}

func (a *auditor) HandleEvent(event gen.MessageEvent) error {
	a.record(event)
	return nil
}

func (a *auditor) record(event gen.MessageEvent) {
	done, ok := event.Message.(EventTaskDone)
	if ok == false {
		a.Log().Warning("unexpected event payload %T", event.Message)
		return
	}
	a.Application().Behavior().(*taskrunnerApp).Record(done.ID, done.Took)
}

func (a *auditor) Terminate(reason error) {
	a.Log().Info("auditor %s stopped: %s", a.PID(), reason)
}
