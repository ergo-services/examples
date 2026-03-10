package messaging

import (
	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

type worker struct {
	act.Actor
}

func factoryWorker() gen.ProcessBehavior {
	return &worker{}
}

func (w *worker) Init(args ...any) error {
	return nil
}

func (w *worker) HandleMessage(from gen.PID, message any) error {
	switch m := message.(type) {
	case MessagePayload:
		w.Log().Debug("received payload %d bytes from %s", len(m.Data), from)
	case MessageBulkPayload:
		w.Log().Debug("received bulk payload %d bytes from %s", len(m.Data), from)
	}
	return nil
}

func (w *worker) Terminate(reason error) {}
