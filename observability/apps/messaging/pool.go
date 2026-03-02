package messaging

import (
	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

type pool struct {
	act.Pool
}

func factoryPool() gen.ProcessBehavior {
	return &pool{}
}

func (p *pool) Init(args ...any) (act.PoolOptions, error) {
	p.Log().Info("messaging pool started on %s", p.Node().Name())
	return act.PoolOptions{
		WorkerFactory: factoryWorker,
		PoolSize:      3,
	}, nil
}

func (p *pool) Terminate(reason error) {
	p.Log().Info("messaging pool terminated: %s", reason)
}
