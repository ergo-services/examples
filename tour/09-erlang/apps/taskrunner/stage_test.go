package taskrunner

import (
	"path/filepath"
	"testing"
	"time"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
	"ergo.services/ergo/testing/check"
	"ergo.services/ergo/testing/stage"
)

// prober lives on the client node: it subscribes to the runner's event and
// submits a task when told to.
type prober struct {
	act.Actor

	runner gen.Atom
}

func factoryProber() gen.ProcessBehavior { return &prober{} }

func (p *prober) Init(args ...any) error {
	p.runner = args[0].(gen.Atom)
	_, err := p.MonitorEvent(Tasks(p.runner))
	return err
}

func (p *prober) HandleMessage(from gen.PID, message any) error {
	if message == "go" {
		return Submit(p, p.runner, "task-1", time.Millisecond)
	}
	return nil
}

// This is the claim the unit harness cannot make: two real nodes, the real
// protocol, a task encoded on the way out and an event encoded on the way back.
func TestTwoNodesTaskAndEvent(t *testing.T) {
	s := stage.New(t)

	runner := s.StartNode("runner", stage.NodeOptions{
		Applications: []gen.ApplicationBehavior{
			CreateApp(Options{Results: filepath.Join(t.TempDir(), "results.log")}),
		},
	})
	client := s.StartNode("client")
	s.Connect(runner, client)

	// Nothing registers the wire types for the client node here, and it still
	// works: the EDF registry is per Go process, and both stage nodes are in
	// this one. Loading the application on the runner registered them for
	// everybody. So this is one mistake the stage cannot catch - a node that
	// forgets RegisterTypes in production fails, and passes here.

	pid := client.Spawn(factoryProber, gen.ProcessOptions{}, runner.Name())
	client.Send(pid, "go")

	// The event came back across the connection.
	client.ShouldReceiveEvent().To(pid).Once().Within(time.Second).Must()

	rec, ok := client.ShouldReceiveEvent().To(pid).Capture()
	check.True(t, ok)

	// The record carries the payload itself, not the gen.MessageEvent wrapper.
	done, ok := rec.Message.(EventTaskDone)
	check.True(t, ok)
	check.Equal(t, "task-1", done.ID)
	check.Equal(t, client.Name(), done.For)
}
