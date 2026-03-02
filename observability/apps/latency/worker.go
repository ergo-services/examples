package latency

import (
	"io"
	"net/http"
	"time"

	"ergo.services/application/radar"
	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

const slowWebURL = "http://slowweb:8080/slow"

type worker struct {
	act.Actor
}

func factoryWorker() gen.ProcessBehavior {
	return &worker{}
}

func (w *worker) Init(args ...any) error {
	if err := radar.RegisterHistogram(w, "slowweb_request_duration_seconds",
		"Duration of HTTP requests to slowweb",
		nil,
		[]float64{0.001, 0.005, 0.01, 0.015, 0.02, 0.05, 0.1}); err != nil {
		w.Log().Error("failed to register histogram: %s", err)
	}

	if err := radar.RegisterCounter(w, "slowweb_requests_total",
		"Total HTTP requests to slowweb",
		[]string{"status"}); err != nil {
		w.Log().Error("failed to register counter: %s", err)
	}

	w.Log().Info("latency worker started on %s", w.Node().Name())
	return nil
}

func (w *worker) HandleMessage(from gen.PID, message any) error {
	switch message.(type) {
	case MessagePing:
		start := time.Now()
		resp, err := http.Get(slowWebURL)
		duration := time.Since(start).Seconds()

		radar.HistogramObserve(w, "slowweb_request_duration_seconds",
			duration, nil)

		if err != nil {
			radar.CounterAdd(w, "slowweb_requests_total",
				1, []string{"error"})
			return nil
		}

		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		radar.CounterAdd(w, "slowweb_requests_total",
			1, []string{"ok"})
	}
	return nil
}

func (w *worker) Terminate(reason error) {
	w.Log().Info("latency worker terminated: %s", reason)
}
