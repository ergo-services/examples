package myapp

import (
	"strings"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
	"ergo.services/ergo/meta"
)

func factory_MyUDP() gen.ProcessBehavior {
	return &MyUDP{}
}

type MyUDP struct {
	act.Actor
}

// Init invoked on a start this process.
func (u *MyUDP) Init(args ...any) error {
	udpOptions := meta.UDPServerOptions{
		Host: "localhost",
		Port: 2345,
		// Process:
		// keep Process empty - all UDP-packets will be handled by this process.
		// for more information, see https://docs.ergo.services/meta-processes/udp
	}

	// create UDP server meta-process (opens UDP-port for the listener)
	metaudp, err := meta.CreateUDPServer(udpOptions)
	if err != nil {
		u.Log().Error("unable to create UDP server meta-process: %s", err)
		return err
	}

	// spawn meta process
	id, err := u.SpawnMeta(metaudp, gen.MetaOptions{})
	if err != nil {
		u.Log().Error("unable to spawn UDP server meta-process: %s", err)
		// we should close listening port
		metaudp.Terminate(err)
		return err
	}

	u.Log().Info("started UDP server on %s:%d (meta-process: %s)", udpOptions.Host, udpOptions.Port, id)
	u.Log().Info("you may check it with command below:")
	u.Log().Info("   $ ncat -u %s %d", udpOptions.Host, udpOptions.Port)
	return nil
}

//
// Methods below are optional, so you can remove those that aren't be used
//

// HandleMessage receives a message with UDP-packet.
func (u *MyUDP) HandleMessage(from gen.PID, message any) error {
	switch m := message.(type) {
	case meta.MessageUDP:
		received := string(m.Data)
		u.Log().Info("got udp packet from %s: %s ", m.Addr, strings.TrimRight(received, "\r\n"))
		m.Data = []byte("OK: " + received)
		if err := u.SendAlias(m.ID, m); err != nil {
			u.Log().Error("unable to send to %s: %s", m.ID, err)
		}
	default:
		u.Log().Info("got unknown message from %s: %#v", from, message)
	}
	return nil
}

// HandleCall invoked if Actor got a synchronous request made with gen.Process.Call(...).
// Return nil as a result to handle this request asynchronously and
// to provide the result later using the gen.Process.SendResponse(...) method.
func (u *MyUDP) HandleCall(from gen.PID, ref gen.Ref, request any) (any, error) {
	u.Log().Info("got request from %s with reference %s", from, ref)
	return gen.Atom("pong"), nil
}

// Terminate invoked on a termination process
func (u *MyUDP) Terminate(reason error) {
	u.Log().Info("terminated with reason: %s", reason)
}

// HandleMessageName invoked if split handling was enabled using SetSplitHandle(true)
// and message has been sent by name
func (u *MyUDP) HandleMessageName(name gen.Atom, from gen.PID, message any) error {
	return nil
}

// HandleMessageAlias invoked if split handling was enabled using SetSplitHandle(true)
// and message has been sent by alias
func (u *MyUDP) HandleMessageAlias(alias gen.Alias, from gen.PID, message any) error {
	return nil
}

// HandleCallName invoked if split handling was enabled using SetSplitHandle(true)
// and request was made by name
func (u *MyUDP) HandleCallName(name gen.Atom, from gen.PID, ref gen.Ref, request any) (any, error) {
	return gen.Atom("pong"), nil
}

// HandleCallAlias invoked if split handling was enabled using SetSplitHandle(true)
// and request was made by alias
func (u *MyUDP) HandleCallAlias(alias gen.Alias, from gen.PID, ref gen.Ref, request any) (any, error) {
	return gen.Atom("pong"), nil
}

// HandleLog invoked on a log message if this process was added as a logger.
// See https://docs.ergo.services/basics/logging for more information
func (u *MyUDP) HandleLog(message gen.MessageLog) error {
	return nil
}

// HandleEvent invoked on an event message if this process got subscribed on
// this event using gen.Process.LinkEvent or gen.Process.MonitorEvent
// See https://docs.ergo.services/basics/events for more information
func (u *MyUDP) HandleEvent(message gen.MessageEvent) error {
	return nil
}

// HandleInspect invoked on the request made with gen.Process.Inspect(...)
func (u *MyUDP) HandleInspect(from gen.PID, item ...string) map[string]string {
	u.Log().Info("got inspect request from %s", from)
	return nil
}
