package myapp

import (
	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

func factory_MyPoolWorker() gen.ProcessBehavior {
	return &MyPoolWorker{}
}

type MyPoolWorker struct {
	act.Actor
}

// Init invoked on a start this process.
func (w *MyPoolWorker) Init(args ...any) error {
	w.Log().Info("started worker process in pool: %s", w.Parent())
	return nil
}

//
// Methods below are optional
//

// HandleMessage invoked if worker process received a message sent with gen.Process.Send(...).
// Non-nil value of the returning error will cause termination of this process.
// To stop this process normally, return gen.TerminateReasonNormal
// or any other for abnormal termination.
// Stopping the worker process causes the spawning of the new worker process by the pool process
func (w *MyPoolWorker) HandleMessage(from gen.PID, message any) error {
	w.Log().Info("worker received message from %s", from)
	return nil
}

// HandleCall invoked if Actor got a synchronous request made with gen.Process.Call(...).
// Return nil as a result to handle this request asynchronously and
// to provide the result later using the gen.Process.SendResponse(...) method.
func (w *MyPoolWorker) HandleCall(from gen.PID, ref gen.Ref, request any) (any, error) {
	w.Log().Info("worker received request from %s with reference %s", from, ref)
	return gen.Atom("pong"), nil
}

// Terminate invoked on a termination process
func (w *MyPoolWorker) Terminate(reason error) {
	w.Log().Info("worker process terminated with reason: %s", reason)
}
