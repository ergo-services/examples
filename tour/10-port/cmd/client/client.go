package main

import (
	"fmt"
	"time"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"

	"taskrunner/step10/apps/taskrunner"
)

type actTick struct{}
type actBurst struct{}

const burst = 60

func factoryClient() gen.ProcessBehavior {
	return &client{}
}

type client struct {
	act.Actor

	runner     gen.Atom
	subscribed bool
	sent       int
	done       int
}

func (c *client) Init(args ...any) error {
	c.runner = args[0].(gen.Atom)

	c.Log().Info("talking to the runner on %s", c.runner)
	if _, err := c.SendAfter(c.PID(), actBurst{}, 30*time.Second); err != nil {
		return err
	}
	return c.Send(c.PID(), actTick{})
}

func (c *client) HandleMessage(from gen.PID, message any) error {
	switch m := message.(type) {

	// one task a second, so there is always something running to look at
	case actTick:
		c.submit()
		_, err := c.SendAfter(c.PID(), actTick{}, time.Second)
		return err

	// and every half a minute, more than the pool can hold
	case actBurst:
		// The tally first, while everything from last time has settled. Right
		// after a burst it would only say that the pool is still working.
		c.Log().Info("so far: %d submitted, %d came back", c.sent, c.done)

		c.Log().Info("--- a burst of %d ---", burst)
		for i := 0; i < burst; i++ {
			c.submit()
		}
		_, err := c.SendAfter(c.PID(), actBurst{}, 30*time.Second)
		return err

	case gen.MessageDownEvent:
		c.subscribed = false
		c.Log().Warning("the event %s is gone: %s", m.Event.Name, m.Reason)
		return nil
	}

	c.Log().Warning("unexpected message %T", message)
	return nil
}

func (c *client) HandleEvent(event gen.MessageEvent) error {
	if _, ok := event.Message.(taskrunner.EventTaskDone); ok {
		c.done++
	}
	return nil
}

func (c *client) submit() {
	c.sent++
	id := fmt.Sprintf("task-%d", c.sent)
	if err := taskrunner.Submit(c, c.runner, id, 150*time.Millisecond); err != nil {
		c.Log().Warning("cannot submit %s: %s", id, err)
		return
	}
	c.subscribe()
}

func (c *client) subscribe() {
	if c.subscribed {
		return
	}
	buffered, err := c.MonitorEvent(taskrunner.Tasks(c.runner))
	if err != nil {
		return
	}
	c.subscribed = true
	c.Log().Info("subscribed to the runner, %d event(s) were waiting", len(buffered))
}
