package taskrunner

import (
	"time"

	"ergo.services/ergo/gen"
)

// The registered names stay private to this package. A caller never learns
// them, so renaming a process breaks nothing outside these four functions.
const (
	supName    gen.Atom = "taskrunner_sup"
	workerName gen.Atom = "taskrunner_worker"
)

// Submit hands a task to the runner. Fire and forget, so it wraps a Send.
func Submit(process gen.Process, id string, effort time.Duration) error {
	return process.Send(workerName, messageSubmit{ID: id, Effort: effort})
}

// Handled asks how many tasks are done. It needs an answer, so it wraps a Call,
// and it hands back a plain int - the response type never leaves this package.
func Handled(process gen.Process) (int, error) {
	result, err := process.Call(workerName, requestStats{})
	if err != nil {
		return 0, err
	}
	return result.(responseStats).Handled, nil
}

// BreakWorker makes the worker fail. It exists for the demo in this step -
// a real application would not offer a way to break itself.
func BreakWorker(process gen.Process) error {
	return process.Send(workerName, messageBreakWorker{})
}

// BreakSupervisor makes the supervisor fail, which is a different thing
// entirely: the supervisor is a member of the application's Group.
func BreakSupervisor(process gen.Process) error {
	return process.Send(supName, messageBreakSupervisor{})
}
