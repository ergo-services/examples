package myapp

import (
	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
	"ergo.services/meta/websocket"
	"fmt"
	"strings"
)

func factory_MyWebSocketWorker() gen.ProcessBehavior {
	return &MyWebSocketWorker{}
}

type MyWebSocketWorker struct {
	act.Actor
}

// Init invoked on a start this process.
func (w *MyWebSocketWorker) Init(args ...any) error {
	w.Log().Info("started WebSocket worker process %s with args %v", w.Name(), args)
	return nil
}

func (w *MyWebSocketWorker) HandleMessage(from gen.PID, message any) error {
	switch m := message.(type) {

	case websocket.MessageConnect:
		w.Log().Info("%s new websocket connection with %s, meta-process %s", w.Name(), m.RemoteAddr.String(), m.ID)
		reply := websocket.Message{
			Body: []byte("hello from " + w.PID().String()),
		}
		w.SendAlias(m.ID, reply)

	case websocket.MessageDisconnect:
		w.Log().Info("%s disconnected with %s", w.Name(), m.ID)

	case websocket.Message:
		received := string(m.Body)
		strip := strings.TrimRight(received, "\r\n")
		w.Log().Info("%s got message (meta-process: %s): %s", w.Name(), m.ID, strip)
		// send echo reply
		reply := fmt.Sprintf("OK %s", strip)
		m.Body = []byte(reply)
		w.SendAlias(m.ID, m)

	default:
		w.Log().Error("uknown message from %s %#v", from, message)
	}
	return nil
}
