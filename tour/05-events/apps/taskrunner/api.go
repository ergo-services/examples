package taskrunner

import (
	"time"

	"ergo.services/ergo/gen"
)

// The registered names stay private to this package. A caller names the node
// the runner is on, and nothing else.
const (
	supName     gen.Atom = "taskrunner_sup"
	workerName  gen.Atom = "taskrunner_worker"
	auditorName gen.Atom = "taskrunner_auditor"
)

// wireTypes are the types that travel between nodes: the task going out, and
// the event coming back.
var wireTypes = []any{messageSubmit{}, EventTaskDone{}}

// RegisterTypes teaches a node the shapes this package puts on the wire. Call
// it right after the node starts and before the first message goes out.
func RegisterTypes(node gen.Node) error {
	return node.Network().RegisterTypes(wireTypes)
}

// Submit hands a task to the runner on the given node. Fire and forget, so it
// wraps a Send.
func Submit(process gen.Process, runner gen.Atom, id string, effort time.Duration) error {
	return process.Send(target(runner), messageSubmit{ID: id, Effort: effort})
}

// target is the only place that knows the shape of an address. A ProcessID is a
// name plus a node, and that node may be the caller's own: the same Send then
// stays local and never touches the network.
func target(runner gen.Atom) gen.ProcessID {
	return gen.ProcessID{Name: workerName, Node: runner}
}
