package main

import (
	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

//
// subscriber (consumer)
//

func factorySub() gen.ProcessBehavior {
	return &actorSub{}
}

type actorSub struct {
	act.Actor
}

func (a *actorSub) Init(args ...any) error {
	// Linking/Monitoring are not allowed here since this
	// process is not fully initialized.
	a.Send(a.PID(), "init")
	a.Log().Info("started subscriber process on: %s", a.Node().Name())
	return nil
}

func (a *actorSub) HandleMessage(from gen.PID, message any) error {
	switch message {
	case "init":
		eventName := gen.Event{
			Name: "myEvent",
			Node: "node-pub@localhost",
		}
		// making subscription using MonitorEvent of the gen.Process interface
		if _, err := a.MonitorEvent(eventName); err != nil {
			return err
		}
		a.Log().Info("successfully subscribed to: %s", eventName)
		return nil
	}
	a.Log().Error("unknown message %#v", message)
	return nil
}

func (a *actorSub) HandleEvent(event gen.MessageEvent) error {
	a.Log().Info("received event %s with value: %#v", event.Event, event.Message)
	return nil
}
