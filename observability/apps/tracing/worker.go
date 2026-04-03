package tracing

import (
	"fmt"
	"math/rand"
	"sort"
	"time"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

// worker periodically triggers traced message chains across the cluster.
// Each tick sends a single message that branches into a deep trace tree.
type worker struct {
	act.Actor

	remotes []gen.Atom
	seq     int
}

func factoryWorker() gen.ProcessBehavior {
	return &worker{}
}

func (w *worker) Init(args ...any) error {
	w.SetTracingSampler(gen.TracingSamplerAlways)
	w.SetTracingAttribute("service", "trace_worker")
	w.SetTracingAttribute("role", "initiator")

	w.Log().Info("tracing worker started on %s", w.Node().Name())
	w.SendAfter(w.PID(), messageTick{}, 3*time.Second)
	return nil
}

func (w *worker) HandleMessage(from gen.PID, message any) error {
	switch message.(type) {
	case messageTick:
		w.discoverRemotes()
		w.seq++
		w.SetTracingSpanAttribute("seq", fmt.Sprintf("%d", w.seq))

		// Each tick sends ONE Call to relay which branches into a deep tree.
		// The relay does: Call sink + Call remote relay + forward pattern.
		// This produces 10-15+ observations in a single trace.
		node := w.pickNode()
		to := gen.ProcessID{Name: relayName, Node: node}
		result, err := w.CallWithTimeout(to, PingRequest{
			Payload: fmt.Sprintf("cycle_%d from %s", w.seq, w.Node().Name()),
		}, 5)
		if err != nil {
			w.Log().Warning("cycle %d failed: %s", w.seq, err)
		} else if resp, ok := result.(PongResponse); ok {
			w.Log().Debug("cycle %d result from %s: %s", w.seq, resp.Node, resp.Payload)
		}

		wait := 3*time.Second + time.Duration(rand.Intn(2000))*time.Millisecond
		w.SendAfter(w.PID(), messageTick{}, wait)
	}
	return nil
}

func (w *worker) discoverRemotes() {
	registrar, err := w.Node().Network().Registrar()
	if err != nil {
		return
	}
	routes, err := registrar.Resolver().ResolveApplication(appName)
	if err != nil {
		return
	}
	myName := w.Node().Name()
	w.remotes = w.remotes[:0]
	for _, route := range routes {
		if route.Node == myName {
			continue
		}
		w.remotes = append(w.remotes, route.Node)
	}
	sort.Slice(w.remotes, func(i, j int) bool {
		return string(w.remotes[i]) < string(w.remotes[j])
	})
}

// pickNode returns local node ~75% of the time, remote ~25%
func (w *worker) pickNode() gen.Atom {
	if len(w.remotes) == 0 || rand.Intn(4) > 0 {
		return w.Node().Name()
	}
	return w.remotes[rand.Intn(len(w.remotes))]
}

func (w *worker) Terminate(reason error) {
	w.Log().Info("tracing worker terminated: %s", reason)
}
