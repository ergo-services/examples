package taskrunner

import (
	"time"

	"ergo.services/ergo/gen"
)

// The registered names stay private to this package. A caller names the node
// the runner is on, and nothing else.
const (
	supName    gen.Atom = "taskrunner_sup"
	workerName gen.Atom = "taskrunner_worker"
)

// wireTypes are the types that travel between nodes. The node hosting the
// application gets them from the spec; a node that only calls the API has to
// register them itself, which is what RegisterTypes is for.
var wireTypes = []any{messageSubmit{}, requestStats{}, responseStats{}}

// RegisterTypes teaches a node the shapes this package puts on the wire. Call
// it right after the node starts and before the first message goes out: a type
// that is not registered cannot be encoded.
func RegisterTypes(node gen.Node) error {
	return node.Network().RegisterTypes(wireTypes)
}

// Submit hands a task to the runner on the given node. Fire and forget, so it
// wraps a Send.
func Submit(process gen.Process, runner gen.Atom, id string, effort time.Duration) error {
	return process.Send(target(runner), messageSubmit{ID: id, Effort: effort})
}

// Handled asks how many tasks are done. It needs an answer, so it wraps a Call,
// and it hands back a plain int - the response type never leaves this package.
func Handled(process gen.Process, runner gen.Atom) (int, error) {
	result, err := process.Call(target(runner), requestStats{})
	if err != nil {
		return 0, err
	}
	return result.(responseStats).Handled, nil
}

// target is the only place that knows the shape of an address. A ProcessID is a
// name plus a node, and that node may be the caller's own: the same Send then
// stays local and never touches the network.
func target(runner gen.Atom) gen.ProcessID {
	return gen.ProcessID{Name: workerName, Node: runner}
}
