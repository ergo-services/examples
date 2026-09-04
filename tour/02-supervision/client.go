package main

import (
	"fmt"
	"time"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

// the acts, driven by the client sending itself a message
type actCrash struct{}
type actReport struct{}
type actStorm struct{}
type actFinal struct{}

func factoryClient() gen.ProcessBehavior {
	return &client{}
}

type client struct {
	act.Actor

	done  int
	storm int
}

func (c *client) Init(args ...any) error {
	// Act one: two ordinary tasks.
	c.submit("task-1", 2)
	c.submit("task-2", 3)
	return nil
}

// submit always addresses the worker by its registered name. That name is what
// survives a restart - the PID behind it does not.
func (c *client) submit(id string, items int) {
	task := Task{ID: id, Effort: 150 * time.Millisecond, Items: items}
	if err := c.Send(gen.Atom("worker"), task); err != nil {
		c.Log().Error("cannot submit %s: %s", id, err)
		return
	}
	c.Log().Info("submitted %s (items: %d)", id, items)
}

func (c *client) HandleMessage(from gen.PID, message any) error {
	switch message.(type) {

	case TaskDone:
		c.done++
		if c.done == 2 {
			_, _ = c.SendAfter(c.PID(), actCrash{}, 300*time.Millisecond)
		}
		return nil

	case actCrash:
		// Act two: a task with no items. The worker divides by zero and dies.
		c.Log().Info("--- one task with Items: 0 ---")
		c.submit("task-bad", 0)
		_, _ = c.SendAfter(c.PID(), actReport{}, 500*time.Millisecond)
		return nil

	case actReport:
		// The name still resolves, so something is answering again.
		result, err := c.Call(gen.Atom("worker"), StatsRequest{})
		if err != nil {
			c.Log().Warning("the worker is gone: %s", err)
			return nil
		}
		stats := result.(StatsResponse)
		c.Log().Info("a worker answers again, tasks it has handled: %d", stats.Handled)
		c.Log().Info("a restart is a new instance, not a resumed one")

		// Act three: keep breaking it, but slowly enough that each bad task
		// reaches a live worker. Crashing faster than the supervisor is willing
		// to restart is the only way to exhaust it.
		c.Log().Info("--- one bad task every 400ms until the supervisor gives up ---")
		_, _ = c.SendAfter(c.PID(), actStorm{}, 400*time.Millisecond)
		return nil

	case actStorm:
		c.storm++
		c.submit(fmt.Sprintf("bad-%d", c.storm), 0)
		if c.storm < 8 {
			_, _ = c.SendAfter(c.PID(), actStorm{}, 400*time.Millisecond)
			return nil
		}
		_, _ = c.SendAfter(c.PID(), actFinal{}, time.Second)
		return nil

	case actFinal:
		if _, err := c.Call(gen.Atom("worker"), StatsRequest{}); err != nil {
			c.Log().Warning("nothing answers any more: %s", err)
			c.Log().Warning("the supervisor exceeded its restart intensity and stopped trying")
		} else {
			c.Log().Info("the worker is still up - it never crashed fast enough")
		}
		c.Log().Info("observer is at http://localhost:9911 - press Ctrl+C to stop")
		return nil
	}

	c.Log().Warning("unexpected message %T", message)
	return nil
}
