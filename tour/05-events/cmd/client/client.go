package main

import (
	"fmt"
	"time"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"

	"taskrunner/step05/apps/taskrunner"
)

type actTick struct{}

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
	return c.Send(c.PID(), actTick{})
}

func (c *client) HandleMessage(from gen.PID, message any) error {
	switch m := message.(type) {

	case actTick:
		c.sent++
		id := fmt.Sprintf("task-%d", c.sent)

		if err := taskrunner.Submit(c, c.runner, id, 150*time.Millisecond); err != nil {
			c.Log().Warning("cannot submit %s: %s", id, err)
		} else {
			c.Log().Info("submitted %s", id)
			c.subscribe()
		}

		_, err := c.SendAfter(c.PID(), actTick{}, time.Second)
		return err

	case gen.MessageDownEvent:
		// The producer went away. Same shape as the node monitor in step 04:
		// the subscription is spent, and the next successful send makes a new
		// one.
		c.subscribed = false
		c.Log().Warning("the event %s is gone: %s", m.Event.Name, m.Reason)
		return nil
	}

	c.Log().Warning("unexpected message %T", message)
	return nil
}

// HandleEvent is where a subscription lands. Nothing here says which process
// published it, and nothing has to: the client subscribed to a name on a node.
func (c *client) HandleEvent(event gen.MessageEvent) error {
	done, ok := event.Message.(taskrunner.EventTaskDone)
	if ok == false {
		c.Log().Warning("unexpected event payload %T", event.Message)
		return nil
	}

	c.done++
	c.Log().Info("%s is done in %s (%d of %d submitted)", done.ID, done.Took, c.done, c.sent)
	return nil
}

// subscribe watches the runner's event stream. Like the node monitor in step
// 04, it needs a live connection, so it waits until a send has gone through.
func (c *client) subscribe() {
	if c.subscribed {
		return
	}

	// MonitorEvent, not LinkEvent: this client has other things to do and
	// should not die with the runner. The buffered messages it returns are the
	// events published before this subscription existed - dropping them here
	// would be losing them.
	buffered, err := c.MonitorEvent(taskrunner.Tasks(c.runner))
	if err != nil {
		c.Log().Warning("cannot subscribe: %s", err)
		return
	}
	c.subscribed = true
	c.Log().Info("subscribed to the runner, %d event(s) were waiting", len(buffered))

	// The buffered ones happened before this client existed - they are somebody
	// else's tasks, or its own from a previous run.
	for _, event := range buffered {
		if done, ok := event.Message.(taskrunner.EventTaskDone); ok {
			c.Log().Info("already done before we arrived: %s in %s", done.ID, done.Took)
		}
	}
}
