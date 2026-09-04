package main

import (
	"time"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"

	"taskrunner/step03/apps/taskrunner"
)

// the acts, driven by the client sending itself a message
type actStats struct{}
type actBreakWorker struct{}
type actAfterWorker struct{}
type actBreakSupervisor struct{}
type actFinal struct{}

func factoryClient() gen.ProcessBehavior {
	return &client{}
}

type client struct {
	act.Actor
}

func (c *client) Init(args ...any) error {
	// Everything below goes through the taskrunner package. The client knows no
	// process name, no message type and no PID - only four functions.
	c.submit("task-1")
	c.submit("task-2")
	c.submit("task-3")

	_, err := c.SendAfter(c.PID(), actStats{}, 800*time.Millisecond)
	return err
}

func (c *client) submit(id string) {
	if err := taskrunner.Submit(c, id, 150*time.Millisecond); err != nil {
		c.Log().Error("cannot submit %s: %s", id, err)
		return
	}
	c.Log().Info("submitted %s", id)
}

func (c *client) HandleMessage(from gen.PID, message any) error {
	switch message.(type) {

	case actStats:
		handled, err := taskrunner.Handled(c)
		if err != nil {
			c.Log().Error("cannot read stats: %s", err)
			return nil
		}
		c.Log().Info("the runner has handled %d tasks", handled)

		c.Log().Info("--- breaking the worker, which is not a group member ---")
		_, _ = c.SendAfter(c.PID(), actBreakWorker{}, 200*time.Millisecond)
		return nil

	case actBreakWorker:
		if err := taskrunner.BreakWorker(c); err != nil {
			c.Log().Error("cannot break the worker: %s", err)
		}
		_, _ = c.SendAfter(c.PID(), actAfterWorker{}, 400*time.Millisecond)
		return nil

	case actAfterWorker:
		// The supervisor put a new worker in place and the application did not
		// take part in any of it. The Call below lands behind task-4 in the same
		// mailbox, so it answers only once that task is finished.
		c.submit("task-4")
		handled, err := taskrunner.Handled(c)
		if err != nil {
			c.Log().Error("cannot read stats: %s", err)
			return nil
		}
		c.Log().Info("the runner answers again, and counts %d - it started over", handled)

		c.Log().Info("--- breaking the supervisor, which is a group member ---")
		_, _ = c.SendAfter(c.PID(), actBreakSupervisor{}, 400*time.Millisecond)
		return nil

	case actBreakSupervisor:
		if err := taskrunner.BreakSupervisor(c); err != nil {
			c.Log().Error("cannot break the supervisor: %s", err)
		}
		_, _ = c.SendAfter(c.PID(), actFinal{}, 600*time.Millisecond)
		return nil

	case actFinal:
		if err := taskrunner.Submit(c, "task-5", time.Millisecond); err != nil {
			c.Log().Warning("the runner is gone: %s", err)
		} else {
			c.Log().Info("the runner is still up")
		}
		if info, err := c.Node().ApplicationInfo("taskrunner"); err == nil {
			c.Log().Info("the application is %s again and can be started once more, "+
				"the node did not go down with it", info.State)
		}
		c.Log().Info("results are in results.log")
		c.Log().Info("observer is at http://localhost:9911 - press Ctrl+C to stop")
		return nil
	}

	c.Log().Warning("unexpected message %T", message)
	return nil
}
