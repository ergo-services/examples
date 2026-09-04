package taskrunner

import (
	"errors"
	"time"

	"ergo.services/ergo/gen"
)

// The registered names stay private to this package. workerName is the pool
// now, and no caller can tell.
const (
	supName     gen.Atom = "taskrunner_sup"
	workerName  gen.Atom = "taskrunner_worker"
	auditorName gen.Atom = "taskrunner_auditor"
)

// wireTypes are the types that travel between nodes.
var wireTypes = []any{messageSubmit{}, EventTaskDone{}, requestAddWorkers{}, responseAddWorkers{}}

// RegisterTypes teaches a node the shapes this package puts on the wire. Call
// it right after the node starts and before the first message goes out.
func RegisterTypes(node gen.Node) error {
	return node.Network().RegisterTypes(wireTypes)
}

// Submit hands a task to the runner on the given node. Exactly the same call as
// in step 05, and it now reaches one of several workers.
func Submit(process gen.Process, runner gen.Atom, id string, effort time.Duration) error {
	return process.Send(target(runner), messageSubmit{ID: id, Effort: effort})
}

// AddWorkers grows the pool and returns its new size.
//
// The priority is what makes this reach the pool at all. Everything arriving at
// normal priority lands in the pool's Main queue, and the pool forwards that
// queue to a worker without looking at it - a request included. High priority
// lands in the System queue, which the pool handles itself.
func AddWorkers(process gen.Process, runner gen.Atom, n int) (int, error) {
	result, err := process.CallWithPriority(
		target(runner), requestAddWorkers{N: n}, gen.MessagePriorityHigh)
	if err != nil {
		return 0, err
	}
	response := result.(responseAddWorkers)
	if response.Error != "" {
		return 0, errors.New(response.Error)
	}
	return response.Size, nil
}

// target is the only place that knows the shape of an address. A ProcessID is a
// name plus a node, and that node may be the caller's own: the same Send then
// stays local and never touches the network.
func target(runner gen.Atom) gen.ProcessID {
	return gen.ProcessID{Name: workerName, Node: runner}
}
