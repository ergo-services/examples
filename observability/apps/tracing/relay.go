package tracing

import (
	"fmt"
	"math/rand"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

// relay handles Call requests and creates deep trace trees by doing
// multiple operations per handler: Call sink, Send sink, Call another relay.
type relay struct {
	act.Actor
}

func factoryRelay() gen.ProcessBehavior {
	return &relay{}
}

func (r *relay) Init(args ...any) error {
	r.SetTracingSampler(gen.TracingSamplerAlways)
	r.SetTracingAttribute("service", "trace_relay")
	r.SetTracingAttribute("role", "transit")
	r.Log().Info("tracing relay started on %s", r.Node().Name())
	return nil
}

func (r *relay) HandleMessage(from gen.PID, message any) error {
	switch msg := message.(type) {
	case MessagePing:
		r.SetTracingSpanAttribute("action", "relay_forward")
		r.SetTracingSpanAttribute("seq", fmt.Sprintf("%d", msg.Seq))
		// forward to local sink
		r.Send(gen.ProcessID{Name: sinkName, Node: r.Node().Name()}, msg)
	}
	return nil
}

func (r *relay) HandleCall(from gen.PID, ref gen.Ref, request any) (any, error) {
	switch req := request.(type) {
	case PingRequest:
		// reserve: synchronous Call to the local sink
		reserve := r.StartTracingSpan("reserve")
		sinkResult, err := r.Call(
			gen.ProcessID{Name: sinkName, Node: r.Node().Name()},
			req,
		)
		if err != nil {
			reserve.EndError(err)
			return PongResponse{Payload: "sink error: " + err.Error(), Node: string(r.Node().Name())}, nil
		}
		reserve.End()

		// validate: synchronous validation Call
		validate := r.StartTracingSpan("validate")
		r.Call(gen.ProcessID{Name: sinkName, Node: r.Node().Name()},
			ValidateRequest{OrderID: fmt.Sprintf("ORD-%d", rand.Intn(10000)), Amount: rand.Intn(500)})
		validate.End()

		// notify: async fan-out to the sink
		notify := r.StartTracingSpan("notify")
		r.Send(gen.ProcessID{Name: sinkName, Node: r.Node().Name()},
			MessageNotify{Kind: "relay_processed", Payload: req.Payload})
		r.Send(gen.ProcessID{Name: sinkName, Node: r.Node().Name()},
			MessageStatus{Service: "relay", Status: "processing"})
		notify.End()

		payload := fmt.Sprintf("relay(%v)", sinkResult)

		// forward to a peer relay to add more depth (~80% chance)
		remotes := r.findRemotes()
		if len(remotes) > 0 && rand.Intn(5) < 4 {
			target := remotes[rand.Intn(len(remotes))]
			forward := r.StartTracingSpan("forward-to-peer")
			forward.SetAttribute("peer", string(target))
			nestedResult, err := r.CallWithTimeout(
				gen.ProcessID{Name: relayName, Node: target},
				MessageForward{
					OriginalSender: string(r.Node().Name()),
					Hops:           []string{string(r.Node().Name()), string(target)},
					Payload:        req.Payload,
				}, 4)
			if err == nil {
				if resp, ok := nestedResult.(PongResponse); ok {
					payload = fmt.Sprintf("relay(%v)+nested(%s)", sinkResult, resp.Payload)
				}
				forward.End()
			} else {
				forward.EndError(err)
			}
		}

		return PongResponse{
			Payload: payload,
			Node:    string(r.Node().Name()),
		}, nil

	case MessageForward:
		// forward-endpoint: the whole forwarded operation as one span
		endpoint := r.StartTracingSpan("forward-endpoint")
		endpoint.SetAttribute("hops", fmt.Sprintf("%d", len(req.Hops)))

		sinkResult, err := r.Call(
			gen.ProcessID{Name: sinkName, Node: r.Node().Name()},
			PingRequest{Payload: req.Payload},
		)
		if err != nil {
			endpoint.EndError(err)
			return PongResponse{Payload: "forward sink error", Node: string(r.Node().Name())}, nil
		}

		// validate
		r.Call(gen.ProcessID{Name: sinkName, Node: r.Node().Name()},
			ValidateRequest{OrderID: fmt.Sprintf("FWD-%d", rand.Intn(10000)), Amount: rand.Intn(500)})

		// notify + status
		r.Send(gen.ProcessID{Name: sinkName, Node: r.Node().Name()},
			MessageNotify{Kind: "forward_complete", Payload: req.Payload})
		r.Send(gen.ProcessID{Name: sinkName, Node: r.Node().Name()},
			MessageStatus{Service: "relay", Status: fmt.Sprintf("forwarded_%d_hops", len(req.Hops))})

		endpoint.End()
		return PongResponse{
			Payload: fmt.Sprintf("forwarded(%v, hops=%d)", sinkResult, len(req.Hops)),
			Node:    string(r.Node().Name()),
		}, nil
	}
	return nil, nil
}

func (r *relay) findRemotes() []gen.Atom {
	registrar, err := r.Node().Network().Registrar()
	if err != nil {
		return nil
	}
	routes, err := registrar.Resolver().ResolveApplication(appName)
	if err != nil {
		return nil
	}
	myName := r.Node().Name()
	var remotes []gen.Atom
	for _, route := range routes {
		if route.Node == myName {
			continue
		}
		remotes = append(remotes, route.Node)
	}
	return remotes
}

func (r *relay) Terminate(reason error) {
	r.Log().Info("tracing relay terminated: %s", reason)
}
