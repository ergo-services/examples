package forest

import (
	"math/rand"
	"time"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

type computeSupervisor struct {
	act.Supervisor
}

func factoryComputeSup() gen.ProcessBehavior {
	return &computeSupervisor{}
}

func (s *computeSupervisor) Init(args ...any) (act.SupervisorSpec, error) {
	return act.SupervisorSpec{
		Type: act.SupervisorTypeAllForOne,
		Restart: act.SupervisorRestart{
			Strategy:  act.SupervisorStrategyPermanent,
			Intensity: 5,
			Period:    10,
		},
		Children: []act.SupervisorChildSpec{
			{Name: computePool, Factory: factoryComputePool},
			{Name: computeCoordinator, Factory: factoryComputeCoordinator},
		},
	}, nil
}

type computePoolProcess struct {
	act.Pool
}

func factoryComputePool() gen.ProcessBehavior {
	return &computePoolProcess{}
}

func (p *computePoolProcess) Init(args ...any) (act.PoolOptions, error) {
	return act.PoolOptions{
		WorkerFactory: factoryComputeWorker,
		PoolSize:      4,
	}, nil
}

type computeWorker struct {
	act.Actor
}

func factoryComputeWorker() gen.ProcessBehavior {
	return &computeWorker{}
}

func (w *computeWorker) HandleMessage(from gen.PID, message any) error {
	switch m := message.(type) {
	case MessageCompute:
		w.Log().Debug("compute worker processed seq=%d size=%d", m.Seq, len(m.Payload))
	}
	return nil
}

type computeCoordinatorProc struct {
	act.Actor
	seq uint64
}

type messageComputeTick struct{}

func factoryComputeCoordinator() gen.ProcessBehavior {
	return &computeCoordinatorProc{}
}

func (c *computeCoordinatorProc) Init(args ...any) error {
	c.SendAfter(c.PID(), messageComputeTick{}, time.Second)
	return nil
}

func (c *computeCoordinatorProc) HandleMessage(from gen.PID, message any) error {
	switch message.(type) {
	case messageComputeTick:
		burst := 5 + rand.Intn(15)
		for i := 0; i < burst; i++ {
			c.seq++
			payload := make([]byte, 64+rand.Intn(512))
			c.Send(computePool, MessageCompute{Seq: c.seq, Payload: payload})
		}
		c.SendAfter(c.PID(), messageComputeTick{}, time.Duration(500+rand.Intn(1500))*time.Millisecond)
	}
	return nil
}
