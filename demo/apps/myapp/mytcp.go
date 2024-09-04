package myapp

import (
	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
	"ergo.services/ergo/meta"
	"strings"
)

func factory_MyTCP() gen.ProcessBehavior {
	return &MyTCP{}
}

type MyTCP struct {
	act.Actor
}

// Init invoked on a start this process.
func (t *MyTCP) Init(args ...any) error {
	var options meta.TCPServerOptions
	options.Port = 12345
	options.Host = "localhost"

	// we don't use options.ProcessPool so all TCP-connections will be handled by this process.
	// for more information, see https://docs.ergo.services/meta-processes/tcp

	// create TCP server meta-process (opens TCP-port for the listener)
	metatcp, err := meta.CreateTCPServer(options)
	if err != nil {
		t.Log().Error("unable to create TCP server: %s", err)
		return err
	}

	// spawn meta process
	id, err := t.SpawnMeta(metatcp, gen.MetaOptions{})
	if err != nil {
		t.Log().Error("unable to spawn TCP server meta-process: %s", err)
		// we should close listening port
		metatcp.Terminate(err)
		return err
	}

	t.Log().Info("started TCP server on %s:%d (meta-process: %s)", options.Host, options.Port, id)
	t.Log().Info("you may check it with command below:")
	t.Log().Info("   $ ncat %s %d", options.Host, options.Port)
	return nil
}

//
// Methods below are optional, so you can remove those that aren't be used
//

// HandleMessage receives a message on a new connection, data packet, or disconnection.
// To serve the new TCP connection, the meta-process of the TCP server spawns the new meta-process.
func (t *MyTCP) HandleMessage(from gen.PID, message any) error {
	switch m := message.(type) {
	case meta.MessageTCPConnect:
		t.Log().Info("new connection with: %s (serving meta-process: %s)", m.RemoteAddr, m.ID)
	case meta.MessageTCPDisconnect:
		t.Log().Info("terminated connection (serving meta-process: %s)", m.ID)
	case meta.MessageTCP:
		received := string(m.Data)
		t.Log().Info("got tcp packet from %s: %s ", m.ID, strings.TrimRight(received, "\r\n"))
		m.Data = []byte("OK: " + received)
		// To write the data to the socket, we should send it as a message
		// to the meta-process that handles this connection
		if err := t.SendAlias(m.ID, m); err != nil {
			t.Log().Error("unable to send to %s: %s", m.ID, err)
		}
	default:
		t.Log().Info("got unknown message from %s: %#v", from, message)
	}
	return nil
}

// HandleCall invoked if Actor got a synchronous request made with gen.Process.Call(...).
// Return nil as a result to handle this request asynchronously and
// to provide the result later using the gen.Process.SendResponse(...) method.
func (t *MyTCP) HandleCall(from gen.PID, ref gen.Ref, request any) (any, error) {
	t.Log().Info("got request from %s with reference %s", from, ref)
	return gen.Atom("pong"), nil
}

// Terminate invoked on a termination process
func (t *MyTCP) Terminate(reason error) {
	t.Log().Info("terminated with reason: %s", reason)
}

// HandleMessageName invoked if split handling was enabled using SetSplitHandle(true)
// and message has been sent by name
func (t *MyTCP) HandleMessageName(name gen.Atom, from gen.PID, message any) error {
	return nil
}

// HandleMessageAlias invoked if split handling was enabled using SetSplitHandle(true)
// and message has been sent by alias
func (t *MyTCP) HandleMessageAlias(alias gen.Alias, from gen.PID, message any) error {
	return nil
}

// HandleCallName invoked if split handling was enabled using SetSplitHandle(true)
// and request was made by name
func (t *MyTCP) HandleCallName(name gen.Atom, from gen.PID, ref gen.Ref, request any) (any, error) {
	return gen.Atom("pong"), nil
}

// HandleCallAlias invoked if split handling was enabled using SetSplitHandle(true)
// and request was made by alias
func (t *MyTCP) HandleCallAlias(alias gen.Alias, from gen.PID, ref gen.Ref, request any) (any, error) {
	return gen.Atom("pong"), nil
}

// HandleLog invoked on a log message if this process was added as a logger.
// See https://docs.ergo.services/basics/logging for more information
func (t *MyTCP) HandleLog(message gen.MessageLog) error {
	return nil
}

// HandleEvent invoked on an event message if this process got subscribed on
// this event using gen.Process.LinkEvent or gen.Process.MonitorEvent
// See https://docs.ergo.services/basics/events for more information
func (t *MyTCP) HandleEvent(message gen.MessageEvent) error {
	return nil
}

// HandleInspect invoked on the request made with gen.Process.Inspect(...)
func (t *MyTCP) HandleInspect(from gen.PID, item ...string) map[string]string {
	t.Log().Info("got inspect request from %s", from)
	return nil
}
