package main

import (
	"fmt"
	"time"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

func factoryClient() gen.ProcessBehavior {
	return &client{}
}

// client submits a few tasks, waits for every answer, then asks the worker what
// it has done. It is an actor too: in Ergo the thing that talks to an actor is
// usually another actor.
type client struct {
	act.Actor

	expected int
	received int
}

func (c *client) Init(args ...any) error {
	c.expected = 5

	for i := 1; i <= c.expected; i++ {
		task := Task{
			ID:     fmt.Sprintf("task-%d", i),
			Effort: time.Duration(i*120) * time.Millisecond,
		}
		// Send is asynchronous: it returns as soon as the message is in the
		// worker's mailbox, not when the work is done.
		if err := c.Send(gen.Atom("worker"), task); err != nil {
			return err
		}
		c.Log().Info("submitted %s", task.ID)
	}

	return nil
}

func (c *client) HandleMessage(from gen.PID, message any) error {
	switch done := message.(type) {
	case TaskDone:
		c.received++
		c.Log().Info("got result for %s (%d of %d)", done.ID, c.received, c.expected)

		if c.received < c.expected {
			return nil
		}

		// Everything is back. Call blocks this actor until the worker answers,
		// which is what makes it the right tool for a question.
		result, err := c.Call(gen.Atom("worker"), StatsRequest{})
		if err != nil {
			return err
		}
		stats := result.(StatsResponse)
		c.Log().Info("all %d tasks done, the worker was busy for %s",
			stats.Handled, stats.Busy.Round(time.Millisecond))
		c.Log().Info("observer is at http://localhost:9911 - press Ctrl+C to stop")
		return nil
	}

	c.Log().Warning("unexpected message %T", message)
	return nil
}
