package main

import (
	"time"

	"ergo.services/ergo/gen"
)

// Task is a unit of work handed to a worker. Effort is how long the work takes,
// which is all the "work" this path ever does.
type Task struct {
	ID     string
	Effort time.Duration
}

// TaskDone is what the worker reports back after finishing one.
type TaskDone struct {
	ID     string
	Worker gen.PID
	Took   time.Duration
}

// StatsRequest asks the worker what it has handled so far. A request expects an
// answer, which is why it is a Call and not a Send - and why the naming differs:
// XXXRequest/XXXResponse for synchronous, MessageXXX for asynchronous.
type StatsRequest struct{}

// StatsResponse is the answer to StatsRequest.
type StatsResponse struct {
	Handled int
	Busy    time.Duration
}
