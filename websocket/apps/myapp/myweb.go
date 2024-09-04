package myapp

import (
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
	act.Actor
}

// Init invoked on a start this process.
func (w *MyWeb) Init(args ...any) error {

	mux := http.NewServeMux()

	// processes to handle messages from the websocket connections
	ws_workers := []gen.Atom{
		"ws_worker_1",
		"ws_worker_2",
		"ws_worker_3",
	}

	// create and spawn websocket handler meta-process
	wsopt := websocket.HandlerOptions{
		ProcessPool: ws_workers,
		// uncomment option below to ignore the origin
		// CheckOrigin: func(r *http.Request) bool { return true },
	}
	wshandler := websocket.CreateHandler(wsopt)
	wshandlerid, err := w.SpawnMeta(wshandler, gen.MetaOptions{})
	if err != nil {
		w.Log().Error("unable to spawn WebSocket WebHandler meta-process: %s", err)
		return err
	}
	mux.Handle("/", wshandler)
	w.Log().Info("started WebSocket handler to serve '/' (meta-process: %s)", wshandlerid)

	// create and spawn web server meta-process
	serverOptions := meta.WebServerOptions{
		Port: 9898,
		Host: "localhost",
		// use node's certificate if it was enabled there
		CertManager: w.Node().CertManager(),
		Handler:     mux,
	}

	webserver, err := meta.CreateWebServer(serverOptions)
	if err != nil {
		w.Log().Error("unable to create Web server meta-process: %s", err)
		return err
	}
	webserverid, err := w.SpawnMeta(webserver, gen.MetaOptions{})
	if err != nil {
		// invoke Terminate to close listening socket
		webserver.Terminate(err)
	}

	https := ""
	if serverOptions.CertManager != nil {
		https = "s"
	}
	w.Log().Info("started Web server %s: ws%s://%s:%d/", webserverid, https, serverOptions.Host, serverOptions.Port)
	w.Log().Info("you may check it with command below:")
	w.Log().Info("   $ websocat -k ws%s://%s:%d", https, serverOptions.Host, serverOptions.Port)
	return nil
}
