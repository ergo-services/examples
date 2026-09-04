package taskrunner

import (
	"time"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

func factoryWorker() gen.ProcessBehavior {
	return &worker{}
}

// worker is no longer registered under a name and no longer owns the event. It
// is one of several identical processes the pool spawns, does one task at a
// time, and reports to the auditor.
type worker struct {
	act.Actor
}

func (w *worker) HandleMessage(from gen.PID, message any) error {
	switch m := message.(type) {
	case messageSubmit:
		started := time.Now()
		work(m.Effort)
		took := time.Since(started).Round(time.Millisecond)

		w.Log().Info("finished %s in %s", m.ID, took)
		return w.Send(auditorName, messageDone{
			ID:   m.ID,
			For:  string(from.Node),
			Took: took,
		})
	}
	return nil
}

// work is the same stand-in as in the earlier steps: a computation, not a sleep.
func work(d time.Duration) {
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
	}
}
