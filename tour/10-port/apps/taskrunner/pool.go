package taskrunner

import (
	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

func factoryPool() gen.ProcessBehavior {
	return &pool{}
}

// pool is the process registered under workerName. Callers still send to that
// one name; behind it are several workers, and the pool hands each message to
// one of them in turn.
type pool struct {
	act.Pool
}

func (p *pool) Init(args ...any) (act.PoolOptions, error) {
	options := act.PoolOptions{
		PoolSize:      3,
		WorkerFactory: factoryWorker,

		// A bound on each worker's mailbox. Without it a slow worker just
		// accumulates work and the pool never learns that it is behind.
		WorkerMailboxSize: 10,
	}

	p.Log().Info("pool %s started with %d workers, mailbox %d each",
		p.PID(), options.PoolSize, options.WorkerMailboxSize)
	return options, nil
}

// HandleCall runs only for requests that arrive above normal priority. Anything
// at normal priority sits in the Main queue, which the pool forwards to a
// worker without inspecting it - so a plain Call would reach a worker instead,
// and a worker knows nothing about the pool it belongs to.
func (p *pool) HandleCall(from gen.PID, ref gen.Ref, request any) (any, error) {
	switch r := request.(type) {
	case requestAddWorkers:
		size, err := p.AddWorkers(r.N)
		if err != nil {
			// The second value here is a termination reason, not an answer.
			// Returning the error would kill the pool and leave the caller
			// waiting for a reply that never comes, so a failure goes in the
			// result like everything else.
			p.Log().Error("cannot add %d workers: %s", r.N, err)
			return responseAddWorkers{Error: err.Error()}, nil
		}
		p.Log().Info("pool grown to %d workers", size)
		return responseAddWorkers{Size: int(size)}, nil
	}
	return nil, nil
}
