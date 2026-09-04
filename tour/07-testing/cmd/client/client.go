package main

import (
	"fmt"
	"time"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"

	"taskrunner/step07/apps/taskrunner"
)

// the acts, driven by the client sending itself a message
type actStart struct{}
type actFlood struct{}
type actReport struct{}
type actGrow struct{}

const flood = 60

func factoryClient() gen.ProcessBehavior {
	return &client{}
}

type client struct {
	act.Actor

	runner gen.Atom
	round  int
	sent   int
	done   int
}

func (c *client) Init(args ...any) error {
	c.runner = args[0].(gen.Atom)

	c.Log().Info("talking to the runner on %s", c.runner)
	return c.Send(c.PID(), actStart{})
}

func (c *client) HandleMessage(from gen.PID, message any) error {
	switch m := message.(type) {

	case actStart:
		// One task, to get a connection and a subscription. Then the flood.
		if err := taskrunner.Submit(c, c.runner, "warmup", 150*time.Millisecond); err != nil {
			c.Log().Warning("the runner is not there yet: %s", err)
			_, err := c.SendAfter(c.PID(), actStart{}, time.Second)
			return err
		}
		if err := c.subscribe(); err != nil {
			return err
		}
		_, err := c.SendAfter(c.PID(), actFlood{}, time.Second)
		return err

	case actFlood:
		c.round++
		c.sent, c.done = 0, 0

		c.Log().Info("--- round %d: %d tasks as fast as they will go ---", c.round, flood)
		for i := 1; i <= flood; i++ {
			id := fmt.Sprintf("r%d-task-%d", c.round, i)
			if err := taskrunner.Submit(c, c.runner, id, 150*time.Millisecond); err != nil {
				c.Log().Warning("cannot submit %s: %s", id, err)
				continue
			}
			c.sent++
		}
		c.Log().Info("%d submitted, every Send returned without an error", c.sent)

		_, err := c.SendAfter(c.PID(), actReport{}, 4*time.Second)
		return err

	case actReport:
		c.Log().Info("round %d: %d submitted, %d finished", c.round, c.sent, c.done)
		if c.done < c.sent {
			c.Log().Warning("%d tasks were dropped - the pool is not a queue", c.sent-c.done)
		}

		if c.round == 1 {
			_, err := c.SendAfter(c.PID(), actGrow{}, time.Second)
			return err
		}
		c.Log().Info("observer is at http://localhost:9911 - press Ctrl+C to stop")
		return nil

	case actGrow:
		size, err := taskrunner.AddWorkers(c, c.runner, 5)
		if err != nil {
			c.Log().Error("cannot grow the pool: %s", err)
			return nil
		}
		c.Log().Info("--- pool grown to %d workers, same flood again ---", size)

		_, err = c.SendAfter(c.PID(), actFlood{}, time.Second)
		return err

	case gen.MessageDownEvent:
		c.Log().Warning("the event %s is gone: %s", m.Event.Name, m.Reason)
		return nil
	}

	c.Log().Warning("unexpected message %T", message)
	return nil
}

// HandleEvent counts what actually got done. The client cannot tell from Send
// whether a task was accepted, so this is the only honest count.
func (c *client) HandleEvent(event gen.MessageEvent) error {
	if _, ok := event.Message.(taskrunner.EventTaskDone); ok {
		c.done++
	}
	return nil
}

func (c *client) subscribe() error {
	buffered, err := c.MonitorEvent(taskrunner.Tasks(c.runner))
	if err != nil {
		return err
	}
	c.Log().Info("subscribed to the runner, %d event(s) were waiting", len(buffered))
	return nil
}
