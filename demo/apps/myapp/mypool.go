package myapp

import (
	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

type MyPool struct {
	act.Pool
}

func factory_MyPool() gen.ProcessBehavior {
	return &MyPool{}
}

// Init invoked on a spawn Pool for the initializing.
func (p *MyPool) Init(args ...any) (act.PoolOptions, error) {
	opts := act.PoolOptions{
		WorkerFactory: factory_MyWebWorker,
		PoolSize:      3,
	}

	p.Log().Info("started process pool of MyWebWorker with %d workers", opts.PoolSize)
	return opts, nil
}

//
// Methods below are optional, so you can remove those that aren't be used
//

// HandleMessage invoked if Pool received a message sent with gen.Process.Send(...) and
// with Priority higher than gen.MessagePriorityNormal. Any other messages are forwarded
// to the process from the pool.
// Non-nil value of the returning error will cause termination of this process.
// To stop this process normally, return gen.TerminateReasonNormal
// or any other for abnormal termination.
func (p *MyPool) HandleMessage(from gen.PID, message any) error {
	p.Log().Info("pool got message from %s", from)
	return nil
}

// HandleCall invoked if Pool got a synchronous request made with gen.Process.Call(...) and
// with Priority higher than gen.MessagePriorityNormal. Any other requests are forwarded
// to the process from the pool.
// Return nil as a result to handle this request asynchronously and
// to provide the result later using the gen.Process.SendResponse(...) method.
func (p *MyPool) HandleCall(from gen.PID, ref gen.Ref, request any) (any, error) {
	p.Log().Info("pool got request from %s with reference %s", from, ref)
	return gen.Atom("pong"), nil
}

// Terminate invoked on a termination process
func (p *MyPool) Terminate(reason error) {
	p.Log().Info("pool process terminated with reason: %s", reason)
}
