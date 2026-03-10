package messaging

import (
	crand "crypto/rand"
	"math/rand"
	"sort"
	"time"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

type bulkSender struct {
	act.Actor

	target int
}

func factoryBulkSender() gen.ProcessBehavior {
	return &bulkSender{}
}

func (s *bulkSender) Init(args ...any) error {
	s.Log().Info("bulk sender started on %s", s.Node().Name())
	s.SetCompression(true)
	s.SendAfter(s.PID(), messageBulkBurst{}, startDelay+3*time.Second)
	return nil
}

func (s *bulkSender) HandleMessage(from gen.PID, message any) error {
	switch message.(type) {
	case messageBulkBurst:
		s.doBurst()
		wait := time.Duration(3+rand.Intn(15)) * time.Second
		s.SendAfter(s.PID(), messageBulkBurst{}, wait)
	}
	return nil
}

func (s *bulkSender) doBurst() {
	registrar, err := s.Node().Network().Registrar()
	if err != nil {
		return
	}

	routes, err := registrar.Resolver().ResolveApplication(appName)
	if err != nil {
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

	count := 5 + rand.Intn(20) // 5..24 large messages
	to := gen.ProcessID{Name: poolName, Node: node}
	for i := 0; i < count; i++ {
		// 100KB..300KB -- large enough to trigger both compression and fragmentation
		length := 100000 + rand.Intn(200001)
		data := make([]byte, length)
		crand.Read(data)
		msg := MessageBulkPayload{Data: data}
		if err := s.Send(to, msg); err != nil {
			s.Log().Warning("bulk sender: send error at %d: %s", i, err)
			return
		}
	}

	s.Log().Debug("bulk sender: burst %d large messages -> %s", count, node)
}

func (s *bulkSender) Terminate(reason error) {
	s.Log().Info("bulk sender terminated: %s", reason)
}
