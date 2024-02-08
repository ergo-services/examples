package myapp

import (
	"bytes"
	"encoding/json"
	"net/http"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
	"ergo.services/ergo/meta"

	"ergo.services/meta/websocket"
)

func factory_MyWeb() gen.ProcessBehavior {
	return &MyWeb{}
}

type MyWeb struct {
	act.Web
}

// Init invoked on a start this process.
func (w *MyWeb) Init(args ...any) (act.WebOptions, error) {
	var options act.WebOptions

	options.Port = 9090
	options.Host = "localhost"

	// enabling TLS with node's certificate
	options.CertManager = w.Node().CertManager()

	mux := http.NewServeMux()

	// create and spawn root handler meta-process.
	// we leave meta.WebHandlerOptions.Process empty to handle HTTP-requests by this process
	// within the HandleMessage callback. But you may want to use the process pool.
	// see https://docs.ergo.services/actors/web for more information
	root := meta.CreateWebHandler(meta.WebHandlerOptions{})
	rootid, err := w.SpawnMeta(root, gen.MetaOptions{})
	if err != nil {
		w.Log().Error("unable to start WebHandler meta-process: %s", err)
		return options, err
	}
	// add it to the mux. you can also use middleware functions:
	// mux.Handle("/", middleware(root))
	mux.Handle("/", root)

	// create and spawn websocket handler meta-process.
	// we leave websocket.HandlerOptions.ProcessPool empty to handle websocket-connections
	// by this process.
	// see https://docs.ergo.services/extra-library/meta-processes/websocket for more information
	ws := websocket.CreateHandler(websocket.HandlerOptions{})
	wsid, err := w.SpawnMeta(ws, gen.MetaOptions{})
	if err != nil {
		w.Log().Error("unable to start WebSocket handler meta-process: %s", err)
		return options, err
	}
	// add it to the mux. you can use middleware function here as well:
	// mux.Handle("/", middleware(ws))
	mux.Handle("/ws", ws)

	options.Handler = mux
	if options.CertManager != nil {
		w.Log().Info("started Web server on https://%s:%d/", options.Host, options.Port)
	} else {
		w.Log().Info("started Web server on http://%s:%d/", options.Host, options.Port)
	}
	w.Log().Info("    endpoint  '/'   (meta-process: %s)", rootid)
	w.Log().Info("    websocket '/ws' (meta-process: %s)", wsid)

	return options, nil
}

func (w *MyWeb) HandleMessage(from gen.PID, message any) error {

	switch m := message.(type) {
	//
	// handle websocket messages
	//
	case websocket.MessageConnect:
		w.Log().Info("new connection with %s. meta-process: %s", m.RemoteAddr, m.ID)
	case websocket.MessageDisconnect:
		w.Log().Info("disconnected. meta-process: %s", m.ID)
	case websocket.Message:
		w.Log().Info("got websocket message from %s", m.ID)

		// send reply as a JSON message with information about this process
		m.Body = w.info()
		if err := w.SendAlias(m.ID, m); err != nil {
			w.Log().Error("unable to send: %s", err)
		}
	//
	// handle HTTP-requests
	//
	case meta.MessageWebRequest:
		defer m.Done()
		w.Log().Info("got HTTP request %q", m.Request.URL.Path)
		m.Response.Header().Set("Content-Type", "application/json")
		// response JSON message with information about this process
		m.Response.Write(w.info())
	default:
		w.Log().Info("unhandled message from %s", from)
	}

	return nil
}

func (w *MyWeb) info() []byte {
	var buf bytes.Buffer

	info, _ := w.Info()
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.Encode(info)

	return buf.Bytes()
}
