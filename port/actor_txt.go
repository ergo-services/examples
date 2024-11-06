package main

import (
	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
	"ergo.services/ergo/meta"

	"strings"
)

func factory_ActorPortTxt() gen.ProcessBehavior {
	return &ActorPortTxt{}
}

type ActorPortTxt struct {
	act.Actor
}

// Init invoked on a start this process.
func (t *ActorPortTxt) Init(args ...any) error {
	var options meta.PortOptions

	options.Cmd = "go"
	options.Args = append(options.Args, "run", "./iotxt/main.go")
	options.Tag = "txt"

	metaport, err := meta.CreatePort(options)
	if err != nil {
		t.Log().Error("unable to create Port: %s", err)
		return err
	}

	// spawn meta process
	id, err := t.SpawnMeta(metaport, gen.MetaOptions{})
	if err != nil {
		t.Log().Error("unable to spawn port meta-process: %s", err)
		// we should close listening port
		metaport.Terminate(err)
		return err
	}

	t.Log().Info("started Port (iotxt) %s (meta-process: %s)", options.Cmd, id)
	return nil
}

func (t *ActorPortTxt) HandleMessage(from gen.PID, message any) error {

	switch m := message.(type) {
	case meta.MessagePortStart:
		t.Log().Info("new port with tag %q (serving meta-process: %s)", m.Tag, m.ID)

	case meta.MessagePortTerminate:
		t.Log().Info("terminated port with tag %s (serving meta-process: %s)", m.Tag, m.ID)

	case meta.MessagePortText:
		t.Log().Info("got TXT data (stdout) from %s: %s ", m.ID, strings.TrimRight(m.Text, "\r\n"))

	case meta.MessagePortError:
		received := m.Error
		t.Log().Info("got ERR data (stderr) from %s: %s ", m.ID, received)
	default:
		t.Log().Info("got unknown message from %s: %#v", from, message)
	}
	return nil
}
