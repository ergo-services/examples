package events

import (
	"fmt"
	"math/rand"
	"sort"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

type subscriber struct {
	act.Actor
}

func factorySubscriber() gen.ProcessBehavior {
	return &subscriber{}
}

func (s *subscriber) Init(args ...any) error {
	index := args[0].(int)

	switch {
	case index < 200:
		// local subscribers: monitor 1-3 random active events (evt_0..9)
		s.subscribeLocal(1 + rand.Intn(3))

	case index < 260:
		// remote subscribers: monitor 1-2 random remote active events (evt_0..9)
		s.subscribeLocal(1)
		s.subscribeRemote(1 + rand.Intn(2))

	default:
		// no_publishing subscribers: monitor evt_17..19 (publishers that never publish)
		evtIndex := 17 + rand.Intn(3)
		evt := gen.Event{
			Name: gen.Atom(fmt.Sprintf("evt_%d", evtIndex)),
			Node: s.Node().Name(),
		}
		s.MonitorEvent(evt)
	}

	return nil
}

func (s *subscriber) subscribeLocal(count int) {
	for i := 0; i < count; i++ {
		evtIndex := rand.Intn(10) // evt_0..9
		evt := gen.Event{
			Name: gen.Atom(fmt.Sprintf("evt_%d", evtIndex)),
			Node: s.Node().Name(),
		}
		s.MonitorEvent(evt)
	}
}

func (s *subscriber) subscribeRemote(count int) {
	remotes := s.resolveRemotes()
	if len(remotes) == 0 {
		return
	}
	for i := 0; i < count; i++ {
		evtIndex := rand.Intn(10) // evt_0..9
		node := remotes[rand.Intn(len(remotes))]
		evt := gen.Event{
			Name: gen.Atom(fmt.Sprintf("evt_%d", evtIndex)),
			Node: node,
		}
		s.MonitorEvent(evt)
	}
}

func (s *subscriber) resolveRemotes() []gen.Atom {
	registrar, err := s.Node().Network().Registrar()
	if err != nil {
		return nil
	}
	routes, err := registrar.Resolver().ResolveApplication(appName)
	if err != nil {
		return nil
	}
	myName := s.Node().Name()
	remotes := make([]gen.Atom, 0, len(routes))
	for _, route := range routes {
		if route.Node == myName {
			continue
		}
		remotes = append(remotes, route.Node)
	}
	sort.Slice(remotes, func(i, j int) bool {
		return string(remotes[i]) < string(remotes[j])
	})
	return remotes
}

func (s *subscriber) HandleEvent(message gen.MessageEvent) error {
	s.Log().Debug("received event %s from %s", message.Event.Name, message.Event.Node)
	return nil
}

func (s *subscriber) Terminate(reason error) {}
