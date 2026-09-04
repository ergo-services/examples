package taskrunner

import "time"

// messageSubmit is unexported: callers go through Submit and never construct
// one. Its fields, on the other hand, must be exported - an unexported field
// cannot be encoded, and registration refuses the type outright.
type messageSubmit struct {
	ID     string
	Effort time.Duration
}
