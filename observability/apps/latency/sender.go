package latency

import (
	"math/rand"
	"sort"
	"time"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

const (
	burstSize = 100
	startDelay = 5 * time.Second
)

type sender struct {
	act.Actor

	target int
}

func factorySender() gen.ProcessBehavior {
	return &sender{}
}

func (s *sender) Init(args ...any) error {
	s.Log().Info("latency sender started on %s", s.Node().Name())
	s.SendAfter(s.PID(), messageBurst{}, startDelay)
	return nil
}

func (s *sender) HandleMessage(from gen.PID, message any) error {
	switch message.(type) {
	case messageBurst:
		s.doBurst()
		wait := time.Duration(1+rand.Intn(5)) * time.Second
		s.SendAfter(s.PID(), messageBurst{}, wait)
	}
	return nil
}

func (s *sender) doBurst() {
	registrar, err := s.Node().Network().Registrar()
	if err != nil {
		s.Log().Warning("latency sender: no registrar: %s", err)
		return
	}

	routes, err := registrar.Resolver().ResolveApplication(appName)
	if err != nil {
		s.Log().Warning("latency sender: resolve failed: %s", err)
		return
	}

	myName := s.Node().Name()
	remotes := make([]gen.Atom, 0, len(routes))
	for _, route := range routes {
		if route.Node == myName {
			continue
		}
		remotes = append(remotes, route.Node)
	}

	if len(remotes) == 0 {
		return
	}

	sort.Slice(remotes, func(i, j int) bool {
		return string(remotes[i]) < string(remotes[j])
	})

	node := remotes[s.target%len(remotes)]
	s.target++

	to := gen.ProcessID{Name: workerName, Node: node}
	for i := 0; i < burstSize; i++ {
		if err := s.Send(to, MessagePing{Seq: i}); err != nil {
			s.Log().Warning("latency sender: send error at %d: %s", i, err)
			return
		}
	}

	s.Log().Info("latency sender: %d messages -> %s", burstSize, node)
}

func (s *sender) Terminate(reason error) {
	s.Log().Info("latency sender terminated: %s", reason)
}
