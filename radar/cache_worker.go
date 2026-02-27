package demo

import (
	"math/rand"
	"time"

	"ergo.services/application/radar"
	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

// cacheWorker simulates a cache service.
// Health: registers for liveness probe with a heartbeat.
// Metrics: maintains a "cache_operations" counter.
type cacheWorker struct {
	act.Actor
	cancelHeartbeat gen.CancelFunc
	cancelMetrics   gen.CancelFunc
}

type messageCacheHeartbeat struct{}
type messageCacheMetrics struct{}

func factoryCacheWorker() gen.ProcessBehavior {
	return &cacheWorker{}
}

func (w *cacheWorker) Init(args ...any) error {
	// register health signal: liveness only, heartbeat every 2s (timeout 8s)
	err := radar.RegisterService(w, "cache", radar.ProbeLiveness, 8*time.Second)
	if err != nil {
		return err
	}

	// register a counter metric for cache operations
	err = radar.RegisterCounter(w, "cache_operations_total", "Total cache operations", []string{"operation"})
	if err != nil {
		return err
	}

	// periodic heartbeat
	w.cancelHeartbeat, _ = w.SendAfter(w.PID(), messageCacheHeartbeat{}, 2*time.Second)

	// periodic metric update
	w.cancelMetrics, _ = w.SendAfter(w.PID(), messageCacheMetrics{}, time.Second)

	w.Log().Info("cache_worker: registered health signal 'cache' and counter 'cache_operations_total'")
	return nil
}

func (w *cacheWorker) HandleMessage(from gen.PID, message any) error {
	switch message.(type) {
	case messageCacheHeartbeat:
		radar.Heartbeat(w, "cache")
		w.cancelHeartbeat, _ = w.SendAfter(w.PID(), messageCacheHeartbeat{}, 2*time.Second)

	case messageCacheMetrics:
		// simulate cache hits and misses
		hits := float64(rand.Intn(50) + 10)
		misses := float64(rand.Intn(10))
		radar.CounterAdd(w, "cache_operations_total", hits, []string{"hit"})
		radar.CounterAdd(w, "cache_operations_total", misses, []string{"miss"})
		w.cancelMetrics, _ = w.SendAfter(w.PID(), messageCacheMetrics{}, time.Second)
	}
	return nil
}

func (w *cacheWorker) Terminate(reason error) {
	if w.cancelHeartbeat != nil {
		w.cancelHeartbeat()
	}
	if w.cancelMetrics != nil {
		w.cancelMetrics()
	}
	w.Log().Info("cache_worker: terminated: %s", reason)
}
