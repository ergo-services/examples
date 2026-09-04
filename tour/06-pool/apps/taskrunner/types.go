package taskrunner

import "time"

// messageSubmit is what a caller sends. It arrives at the pool, which forwards
// it to one of its workers untouched.
type messageSubmit struct {
	ID     string
	Effort time.Duration
}

// messageDone is what a worker reports to the auditor. It never leaves the
// node, so it is not on the wire and needs no registration.
type messageDone struct {
	ID   string
	For  string
	Took time.Duration
}

type requestAddWorkers struct{ N int }

// Error carries a failure back to the caller. HandleCall has no slot for one:
// its second return value is a termination reason.
type responseAddWorkers struct {
	Size  int
	Error string
}
