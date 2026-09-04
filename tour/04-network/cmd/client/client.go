package main

import (
	"fmt"
	"time"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"

	"taskrunner/step04/apps/taskrunner"
)

type actTick struct{}

func factoryClient() gen.ProcessBehavior {
	return &client{}
}

type client struct {
	act.Actor

	runner   gen.Atom
	watching bool
	n        int
}

func (c *client) Init(args ...any) error {
	c.runner = args[0].(gen.Atom)

	c.Log().Info("talking to the runner on %s", c.runner)
	return c.Send(c.PID(), actTick{})
}

func (c *client) HandleMessage(from gen.PID, message any) error {
	switch m := message.(type) {

	case actTick:
		c.n++
		id := fmt.Sprintf("task-%d", c.n)

		if err := taskrunner.Submit(c, c.runner, id, 150*time.Millisecond); err != nil {
			// A remote Send can fail for reasons a local one cannot: the node
			// is not up, or not reachable, or does not know that name.
			c.Log().Warning("cannot submit %s: %s", id, err)
		} else {
			c.Log().Info("submitted %s", id)
			c.watch()
		}

		if c.n%3 == 0 {
			// A Call across the network is the same call, with the wire in the
			// middle. It waits for an answer, so it can also time out.
			handled, err := taskrunner.Handled(c, c.runner)
			if err != nil {
				c.Log().Warning("cannot read stats: %s", err)
			} else {
				c.Log().Info("the runner has handled %d tasks", handled)
			}
		}

		_, err := c.SendAfter(c.PID(), actTick{}, time.Second)
		return err

	case gen.MessageDownNode:
		// The monitor is spent once it fires. The next successful send arms a
		// new one, so a node that comes and goes is noticed every time.
		c.watching = false
		c.Log().Warning("the node %s is down - the monitor said so, no timeout was needed", m.Name)
		return nil
	}

	c.Log().Warning("unexpected message %T", message)
	return nil
}

// watch monitors the runner node, so that losing it is an event rather than a
// timeout. It needs a live connection: monitoring a node nothing has reached yet
// fails with "no route", which is why this waits for a send to succeed first.
func (c *client) watch() {
	if c.watching {
		return
	}
	if err := c.MonitorNode(c.runner); err != nil {
		c.Log().Warning("cannot watch %s: %s", c.runner, err)
		return
	}
	c.watching = true
	c.Log().Info("watching %s", c.runner)
}
