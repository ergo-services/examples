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

	token gen.Ref
}

func (w *worker) Init(args ...any) error {
	// The producer owns the event: it registers the name and gets a token that
	// only it can publish with. Buffer keeps the last few messages for whoever
	// subscribes later.
	token, err := w.RegisterEvent(eventName, gen.EventOptions{
		Notify: true,
		Buffer: 10,
	})
	if err != nil {
		return err
	}
	w.token = token

	w.Log().Info("worker %s is ready, publishing %s", w.PID(), eventName)
	return nil
}

func (w *worker) HandleMessage(from gen.PID, message any) error {
	switch m := message.(type) {
	case messageSubmit:
		started := time.Now()
		work(m.Effort)
		took := time.Since(started).Round(time.Millisecond)

		// One publish, no matter who is listening. The worker does not know
		// the auditor exists, and does not know about any other node either.
		w.Log().Info("finished %s for %s in %s", m.ID, from.Node, took)
		return w.SendEvent(eventName, w.token, EventTaskDone{
			ID:   m.ID,
			For:  from.Node,
			Took: took,
		})

	// Notify: true asked for these two. They arrive when the number of
	// subscribers leaves zero and when it returns to zero - not once per
	// subscriber. A producer whose work is expensive can stop doing it here.
	case gen.MessageEventStart:
		w.Log().Info("%s has listeners now", m.Name)

	case gen.MessageEventStop:
		w.Log().Info("%s has no listeners left", m.Name)
	}
	return nil
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
