package main

import (
	"time"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

func factoryWorker() gen.ProcessBehavior {
	return &worker{}
}

// worker owns its state. No mutex guards handled or busy: the actor handles one
// message at a time on its own goroutine, so nothing else can touch them.
type worker struct {
	act.Actor

	handled int
	busy    time.Duration
}

func (w *worker) Init(args ...any) error {
	w.Log().Info("worker %s is ready", w.PID())
	return nil
}

// HandleMessage receives what was sent with Send. Nobody is waiting for an
// answer, so returning nil simply means "carry on".
func (w *worker) HandleMessage(from gen.PID, message any) error {
	switch task := message.(type) {
	case Task:
		started := time.Now()
		work(task.Effort)
		took := time.Since(started)

		w.handled++
		w.busy += took

		w.Log().Info("finished %s in %s", task.ID, took.Round(time.Millisecond))
		return w.Send(from, TaskDone{ID: task.ID, Worker: w.PID(), Took: took})
	}

	w.Log().Warning("unexpected message %T", message)
	return nil
}

// HandleCall receives what was sent with Call. The caller is blocked until this
// returns, so the value returned here is the answer they get.
func (w *worker) HandleCall(from gen.PID, ref gen.Ref, request any) (any, error) {
	switch request.(type) {
	case StatsRequest:
		return StatsResponse{Handled: w.handled, Busy: w.busy}, nil
	}

	w.Log().Warning("unexpected request %T", request)
	return nil, nil
}

func (w *worker) Terminate(reason error) {
	w.Log().Info("worker stopped: %s", reason)
}

// work stands in for whatever a task would really do. It is a computation and
// not a sleep on purpose: a callback holds the mailbox for as long as it runs,
// so the one thing it must never do is hold it while doing nothing.
func work(d time.Duration) {
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
	}
}
