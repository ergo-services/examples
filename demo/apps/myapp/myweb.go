package myapp

import (
	"net/http"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
	"ergo.services/ergo/meta"
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

	// create and spawn root handler meta-process.
	// see https://docs.ergo.services/actors/web for more information
	// use 'mypool' of WebWorker processes to handle requests
	root := meta.CreateWebHandler(meta.WebHandlerOptions{
		Worker: "mypool",
	})
	rootid, err := w.SpawnMeta(root, gen.MetaOptions{})
	if err != nil {
		w.Log().Error("unable to spawn WebHandler meta-process: %s", err)
		return err
	}
	// add it to the mux. you can also use middleware functions:
	// mux.Handle("/", middleware(root))
	mux.Handle("/", root)
	w.Log().Info("started WebHandler to serve '/' (meta-process: %s)", rootid)

	// create and spawn web server meta-process
	serverOptions := meta.WebServerOptions{
		Port: 9090,
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

	https := "http"
	if serverOptions.CertManager != nil {
		https = "https"
	}
	w.Log().Info("started Web server %s: use %s://%s:%d/", webserverid, https, serverOptions.Host, serverOptions.Port)
	w.Log().Info("you may check it with command below:")
	w.Log().Info("   $ curl -k %s://%s:%d", https, serverOptions.Host, serverOptions.Port)
	return nil
}
