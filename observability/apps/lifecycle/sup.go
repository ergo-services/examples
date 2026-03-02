package lifecycle

import (
	"math/rand"
	"time"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

const startDelay = 3 * time.Second

type lifecycleSup struct {
	act.Supervisor
}

type messageStartChildren struct{}

func factoryLifecycleSup() gen.ProcessBehavior {
	return &lifecycleSup{}
}

func (s *lifecycleSup) Init(args ...any) (act.SupervisorSpec, error) {
	s.SendAfter(s.PID(), messageStartChildren{}, startDelay)

	return act.SupervisorSpec{
		Type: act.SupervisorTypeSimpleOneForOne,
		Restart: act.SupervisorRestart{
			Strategy:  act.SupervisorStrategyPermanent,
			Intensity: 1000,
			Period:    5,
		},
		Children: []act.SupervisorChildSpec{
			{
				Name:    childSpec,
				Factory: factoryChild,
			},
		},
	}, nil
}

func (s *lifecycleSup) HandleMessage(from gen.PID, message any) error {
	switch message.(type) {
	case messageStartChildren:
		count := 10 + rand.Intn(91) // 10..100
		s.Log().Info("lifecycle supervisor starting %d children", count)
		for i := 0; i < count; i++ {
			if err := s.StartChild(childSpec); err != nil {
				s.Log().Warning("lifecycle supervisor: start child failed: %s", err)
			}
		}
	}
	return nil
}
