package taskrunner

import "time"

// The message types are unexported on purpose. Callers go through the functions
// in api.go and never construct one, so their shape stays ours to change.
type messageSubmit struct {
	ID     string
	Effort time.Duration
}

type messageBreakWorker struct{}
type messageBreakSupervisor struct{}

type requestStats struct{}
type responseStats struct{ Handled int }
