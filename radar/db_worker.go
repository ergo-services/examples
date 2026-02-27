package demo

import (
	"math/rand"
	"time"

	"ergo.services/application/radar"
	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

// dbWorker simulates a database service.
// Health: registers for liveness and readiness probes with a heartbeat.
// Metrics: maintains a "db_connections" gauge.
type dbWorker struct {
	act.Actor
	cancelHeartbeat gen.CancelFunc
	cancelMetrics   gen.CancelFunc
}

type messageDBHeartbeat struct{}
type messageDBMetrics struct{}

func factoryDBWorker() gen.ProcessBehavior {
	return &dbWorker{}
}

func (w *dbWorker) Init(args ...any) error {
	// register health signal: liveness + readiness, heartbeat every 3s (timeout 10s)
	err := radar.RegisterService(w, "database", radar.ProbeLiveness|radar.ProbeReadiness, 10*time.Second)
	if err != nil {
		return err
	}

	// register a gauge metric for DB connections
	err = radar.RegisterGauge(w, "db_connections", "Number of active database connections", []string{"pool"})
	if err != nil {
		return err
	}

	// periodic heartbeat
	w.cancelHeartbeat, _ = w.SendAfter(w.PID(), messageDBHeartbeat{}, 3*time.Second)

	// periodic metric update
	w.cancelMetrics, _ = w.SendAfter(w.PID(), messageDBMetrics{}, 2*time.Second)

	w.Log().Info("db_worker: registered health signal 'database' and gauge 'db_connections'")
	return nil
}

func (w *dbWorker) HandleMessage(from gen.PID, message any) error {
	switch message.(type) {
	case messageDBHeartbeat:
		radar.Heartbeat(w, "database")
		w.cancelHeartbeat, _ = w.SendAfter(w.PID(), messageDBHeartbeat{}, 3*time.Second)

	case messageDBMetrics:
		// simulate fluctuating connection count
		connections := 5.0 + float64(rand.Intn(20))
		radar.GaugeSet(w, "db_connections", connections, []string{"primary"})
		w.cancelMetrics, _ = w.SendAfter(w.PID(), messageDBMetrics{}, 2*time.Second)
	}
	return nil
}

func (w *dbWorker) Terminate(reason error) {
	if w.cancelHeartbeat != nil {
		w.cancelHeartbeat()
	}
	if w.cancelMetrics != nil {
		w.cancelMetrics()
	}
	w.Log().Info("db_worker: terminated: %s", reason)
}
