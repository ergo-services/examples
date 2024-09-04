package myapp

import (
	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

func factory_MySup() gen.ProcessBehavior {
	return &MySup{}
}

type MySup struct {
	act.Supervisor
}

// Init invoked on a spawn Supervisor process. This is a mandatory callback for the implementation
func (sup *MySup) Init(args ...any) (act.SupervisorSpec, error) {
	var spec act.SupervisorSpec

	// set supervisor type
	spec.Type = act.SupervisorTypeOneForOne

	// add children
	spec.Children = []act.SupervisorChildSpec{
		{
			Name:    "mypool",
			Factory: factory_MyPool,
		},
		{
			Name:    "myweb",
			Factory: factory_MyWeb,
		},
		{
			Name:    "mytcp",
			Factory: factory_MyTCP,
		},
		{
			Name:    "myudp",
			Factory: factory_MyUDP,
		},
	}

	// set strategy
	spec.Restart.Strategy = act.SupervisorStrategyTransient
	spec.Restart.Intensity = 2 // How big bursts of restarts you want to tolerate.
	spec.Restart.Period = 5    // In seconds.

	return spec, nil
}

//
// Methods below are optional, so you can remove those that aren't be used
//

// HandleChildStart invoked on a successful child process starting if option EnableHandleChild
// was enabled in act.SupervisorSpec
func (sup *MySup) HandleChildStart(name gen.Atom, pid gen.PID) error {
	return nil
}

// HandleChildTerminate invoked on a child process termination if option EnableHandleChild
// was enabled in act.SupervisorSpec
func (sup *MySup) HandleChildTerminate(name gen.Atom, pid gen.PID, reason error) error {
	return nil
}

// HandleMessage invoked if Supervisor received a message sent with gen.Process.Send(...).
// Non-nil value of the returning error will cause termination of this process.
// To stop this process normally, return gen.TerminateReasonNormal or
// gen.TerminateReasonShutdown. Any other - for abnormal termination.
func (sup *MySup) HandleMessage(from gen.PID, message any) error {
	sup.Log().Info("supervisor got message from %s", from)
	return nil
}

// HandleCall invoked if Supervisor got a synchronous request made with gen.Process.Call(...).
// Return nil as a result to handle this request asynchronously and
// to provide the result later using the gen.Process.SendResponse(...) method.
func (sup *MySup) HandleCall(from gen.PID, ref gen.Ref, request any) (any, error) {
	sup.Log().Info("supervisor got request from %s with reference %s", from, ref)
	return gen.Atom("pong"), nil
}

// Terminate invoked on a termination supervisor process
func (sup *MySup) Terminate(reason error) {
	sup.Log().Info("supervisor terminated with reason: %s", reason)
}

// HandleInspect invoked on the request made with gen.Process.Inspect(...)
func (sup *MySup) HandleInspect(from gen.PID, item ...string) map[string]string {
	sup.Log().Info("supervisor got inspect request from %s", from)
	return nil
}
