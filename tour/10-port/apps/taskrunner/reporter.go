package taskrunner

import (
	"fmt"
	"strings"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
	"ergo.services/ergo/meta"
)

const reportProgram = "awk"

func factoryReporter() gen.ProcessBehavior {
	return &reporter{}
}

// reporter hands every finished task to an external program and logs what it
// answers back. The program here is awk reading stdin and writing stdout; it
// could as well be Python, a C binary or something written in 1998 - this actor
// only writes lines and reads lines.
type reporter struct {
	act.Actor

	port gen.Alias
}

func (r *reporter) Init(args ...any) error {
	// The port runs the program as a child process, writes what we send it to
	// its stdin, and turns everything it prints into messages to this actor.
	// Text mode is the default, so its output arrives line by line.
	port, err := meta.CreatePort(meta.PortOptions{
		Cmd:  reportProgram,
		Args: []string{"-f", "report.awk"},
		Tag:  "report",
	})
	if err != nil {
		return err
	}

	alias, err := r.SpawnMeta(port, gen.MetaOptions{})
	if err != nil {
		return err
	}
	r.port = alias

	buffered, err := r.MonitorEvent(Tasks(r.Node().Name()))
	if err != nil {
		return err
	}
	for _, event := range buffered {
		if err := r.HandleEvent(event); err != nil {
			return err
		}
	}

	r.Log().Info("reporter %s is ready, feeding %s through %s", r.PID(), reportProgram, alias)
	return nil
}

// HandleEvent turns one finished task into one line on the program's stdin.
func (r *reporter) HandleEvent(event gen.MessageEvent) error {
	done, ok := event.Message.(EventTaskDone)
	if ok == false {
		return nil
	}

	// The same type carries both directions: sent to the port it is a write to
	// stdin, received from the port it is a line of stdout.
	//
	// The error is returned rather than logged. Without the program this actor
	// has no job, so it dies here, and the supervisor puts back both it and a
	// fresh program. Nothing to repair by hand.
	return r.Send(r.port, meta.MessagePortText{
		Text: fmt.Sprintf("%s %d\n", done.ID, done.Took.Milliseconds()),
	})
}

func (r *reporter) HandleMessage(from gen.PID, message any) error {
	switch m := message.(type) {

	case meta.MessagePortText:
		r.Log().Info("report: %s", strings.TrimRight(m.Text, "\r\n"))

	case meta.MessagePortStart:
		r.Log().Info("%s started, port %s", reportProgram, m.ID)

	case meta.MessagePortTerminate:
		r.Log().Info("%s is gone, port %s", reportProgram, m.ID)

	case meta.MessagePortError:
		// Whatever the program wrote to stderr, plus the read errors of the
		// port itself. The program dying shows up here first.
		r.Log().Error("%s: %s", reportProgram, m.Error)

	case gen.MessageDownEvent:
		r.Log().Warning("the event %s is gone, the supervisor will put us back", m.Event.Name)
	}
	return nil
}

func (r *reporter) Terminate(reason error) {
	// Nothing to close: the port is a meta-process of this actor, so it goes
	// when the actor goes, and the external program with it.
	r.Log().Info("reporter %s stopped: %s", r.PID(), reason)
}
