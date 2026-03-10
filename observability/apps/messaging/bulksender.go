package messaging

import (
	"bytes"
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
		// 100KB..300KB -- large enough to trigger fragmentation
		length := 100000 + rand.Intn(200001)
		pattern := []byte("order_id:12345 status:active price:99.95 amount:1.5 exchange:binance ")
		data := bytes.Repeat(pattern, length/len(pattern)+1)[:length]
		// random noise tail (10-80% of data) for varied compression ratio
		noiseStart := length * (1 + rand.Intn(8)) / 10 // 10%..80% structured, rest is noise
		for j := noiseStart; j < length; j++ {
			data[j] = byte(rand.Intn(256))
		}

		// randomly toggle compression per message
		s.SetCompression(rand.Intn(2) == 0)

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
