package main

import (
	"time"

	"ergo.services/ergo/gen"
)

// Task now says how many items it carries. The worker splits the effort over
// them - which is where the bug lives.
type Task struct {
	ID     string
	Effort time.Duration
	Items  int
}

type TaskDone struct {
	ID     string
	Worker gen.PID
	Took   time.Duration
}

type StatsRequest struct{}

type StatsResponse struct {
	Handled int
}
