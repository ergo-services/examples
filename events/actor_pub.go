package main

import (
	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
	"ergo.services/ergo/lib"
	"time"
)

//
// publisher (producer)
//

func factoryPub() gen.ProcessBehavior {
	return &actorPub{}
}

type actorPub struct {
	act.Actor

	token         gen.Ref
	haveConsumers bool
}

func (a *actorPub) Init(args ...any) error {
	a.Log().Info("started publisher process on: %s", a.Node().Name())
	// event registration is not allowed here since this
	// process is not fully initialized.
	a.Send(a.PID(), "init")
	return nil
}

func (a *actorPub) HandleMessage(from gen.PID, message any) error {
	eventName := gen.Atom("myEvent")
	eventStart := gen.MessageEventStart{Name: eventName}
	eventStop := gen.MessageEventStop{Name: eventName}

	switch message {

	case "init": // handle "init" message
		evOptions := gen.EventOptions{
			// enable notification on first/last consumer
			Notify: true,
		}
		token, err := a.RegisterEvent(eventName, evOptions)
		if err != nil {
			return err
		}
		a.token = token
		a.Log().Info("registered event %s, waiting for consumers...", eventName)

		return nil

	case eventStart: // handle gen.MessageEventStart message
		a.Log().Info("publisher got first consumer for %s. start producing events...", eventName)
		a.haveConsumers = true
		a.Send(a.PID(), "produce")
		return nil

	case "produce": // handle "produce" message
		if a.haveConsumers == false {
			// no consumers. ignoring this message
			return nil
		}

		// produce and publish event
		event := MyPubMessage{
			MyString: lib.RandomString(4),
		}
		a.Log().Info("publishing event with value: %#v", event)
		a.SendEvent(eventName, a.token, event)

		// schedule next event in one second
		a.SendAfter(a.PID(), "produce", time.Second)
		return nil

	case eventStop: // handle gen.MessageEventStop message
		a.Log().Info("no consumers for %s", eventName)
		a.haveConsumers = false
		return nil
	}

	a.Log().Error("unknown message %#v", message)
	return nil
}
