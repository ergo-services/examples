package taskrunner

import "time"

// The message types are unexported on purpose. Callers go through the functions
// in api.go and never construct one, and that holds across the network too: the
// wire format is registered by type, and both nodes get the type by importing
// this package.
//
// Every field, on the other hand, must be exported. An unexported field cannot
// be encoded, and registration refuses the type outright.
type messageSubmit struct {
	ID     string
	Effort time.Duration
}

type requestStats struct{}
type responseStats struct{ Handled int }
