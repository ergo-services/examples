package taskrunner

import (
	"time"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

func factoryWorker() gen.ProcessBehavior {
	return &worker{}
}

type worker struct {
	act.Actor

	handled int
}

func (w *worker) Init(args ...any) error {
	w.Log().Info("worker %s is ready", w.PID())
	return nil
}

// HandleMessage does not care where the message came from. A local Send and a
// Send from another node arrive here the same way, and from names the sender.
func (w *worker) HandleMessage(from gen.PID, message any) error {
	switch m := message.(type) {
	case messageSubmit:
		started := time.Now()
		work(m.Effort)
		took := time.Since(started).Round(time.Millisecond)
		w.handled++

		// The results file belongs to the application, not to this process. It
		// was open before this worker existed and it stays open after it dies.
		w.Application().Behavior().(*taskrunnerApp).Record(m.ID, took)
		w.Log().Info("finished %s for %s in %s", m.ID, from.Node, took)
	}
	return nil
}

func (w *worker) HandleCall(from gen.PID, ref gen.Ref, request any) (any, error) {
	switch request.(type) {
	case requestStats:
		return responseStats{Handled: w.handled}, nil
	}
	return nil, nil
}

func (w *worker) Terminate(reason error) {
	w.Log().Info("worker %s stopped: %s", w.PID(), reason)
}

// work is the same stand-in as in the earlier steps: a computation, not a sleep.
func work(d time.Duration) {
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
	}
}
