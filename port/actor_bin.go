package main

import (
	"sync"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
	"ergo.services/ergo/meta"
)

var (
	buffs = &sync.Pool{
		New: func() interface{} {
			return make([]byte, 1024)
		},
	}
)

func factory_ActorPortBin() gen.ProcessBehavior {
	return &ActorPortBin{}
}

type ActorPortBin struct {
	act.Actor
}

// Init invoked on a start this process.
func (t *ActorPortBin) Init(args ...any) error {
	var options meta.PortOptions

	options.Cmd = "go"
	options.Args = append(options.Args, "run", "./iobin/main.go")
	options.Tag = "bin"
	options.Binary.Enable = true
	options.Binary.ChunkHeaderSize = 7
	options.Binary.ChunkHeaderLengthPosition = 3
	options.Binary.ChunkHeaderLengthSize = 4
	options.Binary.ChunkHeaderLengthIncludesHeader = true
	options.ReadBufferPool = buffs // use pool of buffers

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

	t.Log().Info("started Port %s (meta-process: %s)", options.Cmd, id)
	return nil
}

func (t *ActorPortBin) HandleMessage(from gen.PID, message any) error {

	switch m := message.(type) {
	case meta.MessagePortStarted:
		t.Log().Info("new port with tag %q (serving meta-process: %s)", m.Tag, m.ID)

	case meta.MessagePortTerminated:
		t.Log().Info("terminated port with tag %s (serving meta-process: %s)", m.Tag, m.ID)

	case meta.MessagePortData:
		msg := m.Data[7:] // cut the header (defined in options.ChunkHeaderSize)
		t.Log().Info("got BIN data (stdout) from %s: %q ", m.ID, string(msg))
		buffs.Put(m.Data) // put back the buffer into the pool

	case meta.MessagePortError:
		received := m.Error
		t.Log().Info("got ERR data (stderr) from %s: %s ", m.ID, received)
	default:
		t.Log().Info("got unknown message from %s: %#v", from, message)
	}
	return nil
}
