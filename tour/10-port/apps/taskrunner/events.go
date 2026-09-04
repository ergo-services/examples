package taskrunner

import (
	"time"

	"ergo.services/ergo/gen"
)

// eventName is the stream the application publishes on. Subscribers name it
// through Tasks, so it stays private like the process names.
const eventName gen.Atom = "tasks"

// EventTaskDone is published once per finished task.
//
// This one is exported, and that is the difference between an event and a
// message. A message is addressed to a process, so the package can hide it
// behind a function. An event is addressed to nobody in particular: whoever
// subscribes has to be able to name the type to match on it.
type EventTaskDone struct {
	ID   string
	For  gen.Atom
	Took time.Duration
}

// Tasks is the event to subscribe to for task completions on the given node.
func Tasks(runner gen.Atom) gen.Event {
	return gen.Event{Name: eventName, Node: runner}
}
