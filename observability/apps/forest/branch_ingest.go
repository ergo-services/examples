package forest

import (
	"math/rand"
	"time"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

const (
	ingestRouteFast gen.Atom = "forest_ingest_fast"
	ingestRouteSlow gen.Atom = "forest_ingest_slow"
	ingestRouteBulk gen.Atom = "forest_ingest_bulk"
)

type ingestSupervisor struct {
	act.Supervisor
}

func factoryIngestSup() gen.ProcessBehavior {
	return &ingestSupervisor{}
}

func (s *ingestSupervisor) Init(args ...any) (act.SupervisorSpec, error) {
	return act.SupervisorSpec{
		Type: act.SupervisorTypeRestForOne,
		Restart: act.SupervisorRestart{
			Strategy:  act.SupervisorStrategyPermanent,
			Intensity: 5,
			Period:    10,
		},
		Children: []act.SupervisorChildSpec{
			{Name: ingestRouter, Factory: factoryIngestRouter},
			{Name: ingestAggregator, Factory: factoryIngestAggregator},
		},
	}, nil
}

type ingestRouterProc struct {
	act.Router
}

func factoryIngestRouter() gen.ProcessBehavior {
	return &ingestRouterProc{}
}

func (r *ingestRouterProc) Init(args ...any) (act.RouterOptions, error) {
	return act.RouterOptions{
		Routes: []act.Route{
			{Name: ingestRouteFast, Factory: factoryIngestRouteWorker},
			{Name: ingestRouteSlow, Factory: factoryIngestRouteWorker},
			{Name: ingestRouteBulk, Factory: factoryIngestRouteWorker},
		},
	}, nil
}

func (r *ingestRouterProc) RouteMessage(from gen.PID, message any) gen.Atom {
	if m, ok := message.(MessageIngest); ok {
		return m.Stream
	}
	return act.RouteDiscard
}

func (r *ingestRouterProc) RouteCall(from gen.PID, ref gen.Ref, request any) gen.Atom {
	return act.RouteDiscard
}

func (r *ingestRouterProc) HandleMessage(from gen.PID, message any) error {
	return nil
}

func (r *ingestRouterProc) HandleCall(from gen.PID, ref gen.Ref, request any) (any, error) {
	return nil, nil
}

func (r *ingestRouterProc) HandleEvent(message gen.MessageEvent) error {
	return nil
}

func (r *ingestRouterProc) HandleInspect(from gen.PID, item ...string) map[string]string {
	return nil
}

type ingestRouteWorker struct {
	act.Actor
}

func factoryIngestRouteWorker() gen.ProcessBehavior {
	return &ingestRouteWorker{}
}

func (w *ingestRouteWorker) HandleMessage(from gen.PID, message any) error {
	switch m := message.(type) {
	case MessageIngest:
		w.Log().Debug("ingest route worker processed stream=%s value=%d", m.Stream, m.Value)
	}
	return nil
}

type ingestAggregatorProc struct {
	act.Actor
}

type messageIngestTick struct{}

func factoryIngestAggregator() gen.ProcessBehavior {
	return &ingestAggregatorProc{}
}

func (a *ingestAggregatorProc) Init(args ...any) error {
	a.SendAfter(a.PID(), messageIngestTick{}, time.Second)
	return nil
}

func (a *ingestAggregatorProc) HandleMessage(from gen.PID, message any) error {
	switch message.(type) {
	case messageIngestTick:
		streams := []gen.Atom{ingestRouteFast, ingestRouteSlow, ingestRouteBulk}
		for i := 0; i < 10+rand.Intn(20); i++ {
			a.Send(ingestRouter, MessageIngest{
				Stream: streams[rand.Intn(len(streams))],
				Value:  rand.Int63n(10000),
			})
		}
		a.SendAfter(a.PID(), messageIngestTick{}, time.Duration(500+rand.Intn(1500))*time.Millisecond)
	}
	return nil
}
