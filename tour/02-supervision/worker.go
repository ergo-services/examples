package main

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

func (w *worker) HandleMessage(from gen.PID, message any) error {
	switch task := message.(type) {
	case Task:
		// The bug: nothing checks Items. A task with none divides by zero and
		// the worker panics - the ordinary kind of mistake supervision is for.
		perItem := task.Effort / time.Duration(task.Items)

		started := time.Now()
		work(perItem * time.Duration(task.Items))
		took := time.Since(started)

		w.handled++
		w.Log().Info("finished %s in %s (handled %d since start)",
			task.ID, took.Round(time.Millisecond), w.handled)

		return w.Send(from, TaskDone{ID: task.ID, Worker: w.PID(), Took: took})
	}

	w.Log().Warning("unexpected message %T", message)
	return nil
}

func (w *worker) HandleCall(from gen.PID, ref gen.Ref, request any) (any, error) {
	switch request.(type) {
	case StatsRequest:
		return StatsResponse{Handled: w.handled}, nil
	}
	return nil, nil
}

func (w *worker) Terminate(reason error) {
	w.Log().Info("worker %s stopped: %s", w.PID(), reason)
}

// work is the same stand-in as in step 01: a computation, not a sleep.
func work(d time.Duration) {
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
	}
}
