package events

import (
	"time"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

const (
	numPublishers  = 20
	numSubscribers = 300

	publisherSpec  gen.Atom = "events_publisher"
	subscriberSpec gen.Atom = "events_subscriber"
)

type eventsSup struct {
	act.Supervisor
}

func factoryEventsSup() gen.ProcessBehavior {
	return &eventsSup{}
}

func (s *eventsSup) Init(args ...any) (act.SupervisorSpec, error) {
	s.SendAfter(s.PID(), messageStartPublishers{}, 1*time.Second)

	return act.SupervisorSpec{
		Type: act.SupervisorTypeSimpleOneForOne,
		Restart: act.SupervisorRestart{
			Strategy:  act.SupervisorStrategyPermanent,
			Intensity: 500,
			Period:    5,
		},
		Children: []act.SupervisorChildSpec{
			{
				Name:    publisherSpec,
				Factory: factoryPublisher,
			},
			{
				Name:    subscriberSpec,
				Factory: factorySubscriber,
			},
		},
	}, nil
}

func (s *eventsSup) HandleMessage(from gen.PID, message any) error {
	switch message.(type) {
	case messageStartPublishers:
		s.Log().Info("events: starting %d publishers", numPublishers)
		for i := 0; i < numPublishers; i++ {
			if err := s.StartChild(publisherSpec, i); err != nil {
				s.Log().Warning("events: start publisher %d failed: %s", i, err)
			}
		}
		// delay subscribers to let all nodes register events
		s.SendAfter(s.PID(), messageStartSubscribers{}, 5*time.Second)

	case messageStartSubscribers:
		s.Log().Info("events: starting %d subscribers", numSubscribers)
		for i := 0; i < numSubscribers; i++ {
			if err := s.StartChild(subscriberSpec, i); err != nil {
				s.Log().Warning("events: start subscriber %d failed: %s", i, err)
			}
		}
	}
	return nil
}
