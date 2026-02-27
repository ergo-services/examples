package demo

import (
	"math/rand"
	"time"

	"ergo.services/application/radar"
	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

// apiWorker simulates an API gateway service.
// Health: registers for readiness and startup probes (no heartbeat, uses manual up/down).
// Metrics: maintains a "http_request_duration_seconds" histogram.
type apiWorker struct {
	act.Actor
	cancelMetrics gen.CancelFunc
}

type messageAPIMetrics struct{}

func factoryAPIWorker() gen.ProcessBehavior {
	return &apiWorker{}
}

func (w *apiWorker) Init(args ...any) error {
	// register health signal: readiness + startup, no heartbeat timeout
	err := radar.RegisterService(w, "api_gateway", radar.ProbeReadiness|radar.ProbeStartup, 0)
	if err != nil {
		return err
	}

	// register a histogram for request durations
	buckets := []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5}
	err = radar.RegisterHistogram(w, "http_request_duration_seconds",
		"HTTP request latency in seconds", []string{"method", "path"}, buckets)
	if err != nil {
		return err
	}

	// mark as up immediately (no heartbeat needed, signal stays up until we say otherwise)
	radar.ServiceUp(w, "api_gateway")

	// periodic metric update simulating request latencies
	w.cancelMetrics, _ = w.SendAfter(w.PID(), messageAPIMetrics{}, 500*time.Millisecond)

	w.Log().Info("api_worker: registered health signal 'api_gateway' and histogram 'http_request_duration_seconds'")
	return nil
}

func (w *apiWorker) HandleMessage(from gen.PID, message any) error {
	switch message.(type) {
	case messageAPIMetrics:
		// simulate some request latencies
		methods := []string{"GET", "POST"}
		paths := []string{"/api/users", "/api/orders", "/api/products"}

		method := methods[rand.Intn(len(methods))]
		path := paths[rand.Intn(len(paths))]

		// random latency between 1ms and 500ms
		latency := 0.001 + rand.Float64()*0.5
		radar.HistogramObserve(w, "http_request_duration_seconds", latency, []string{method, path})

		w.cancelMetrics, _ = w.SendAfter(w.PID(), messageAPIMetrics{}, 500*time.Millisecond)
	}
	return nil
}

func (w *apiWorker) Terminate(reason error) {
	if w.cancelMetrics != nil {
		w.cancelMetrics()
	}
	w.Log().Info("api_worker: terminated: %s", reason)
}
