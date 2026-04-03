package tracing

import (
	"fmt"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

// sink receives messages and calls from relay and worker.
// Handles various message types to create diverse trace observations.
type sink struct {
	act.Actor

	received int
}

func factorySink() gen.ProcessBehavior {
	return &sink{}
}

func (s *sink) Init(args ...any) error {
	s.SetTracingSampler(gen.TracingSamplerAlways)
	s.SetTracingAttribute("service", "trace_sink")
	s.SetTracingAttribute("role", "receiver")
	s.Log().Info("tracing sink started on %s", s.Node().Name())
	return nil
}

func (s *sink) HandleMessage(from gen.PID, message any) error {
	s.received++
	switch msg := message.(type) {
	case MessagePing:
		s.SetTracingSpanAttribute("action", "ping_received")
		s.SetTracingSpanAttribute("seq", fmt.Sprintf("%d", msg.Seq))
	case MessageNotify:
		s.SetTracingSpanAttribute("action", "notify_received")
		s.SetTracingSpanAttribute("kind", msg.Kind)
	case MessageStatus:
		s.SetTracingSpanAttribute("action", "status_received")
		s.SetTracingSpanAttribute("service", msg.Service)
		s.SetTracingSpanAttribute("status", msg.Status)
	}
	s.SetTracingSpanAttribute("total_received", fmt.Sprintf("%d", s.received))
	return nil
}

func (s *sink) HandleCall(from gen.PID, ref gen.Ref, request any) (any, error) {
	s.received++
	switch req := request.(type) {
	case PingRequest:
		s.SetTracingSpanAttribute("action", "ping_call")
		s.SetTracingSpanAttribute("total_received", fmt.Sprintf("%d", s.received))
		return PongResponse{
			Payload: "sink_ack: " + req.Payload,
			Node:    string(s.Node().Name()),
		}, nil

	case ValidateRequest:
		s.SetTracingSpanAttribute("action", "validate")
		s.SetTracingSpanAttribute("order_id", req.OrderID)
		s.SetTracingSpanAttribute("amount", fmt.Sprintf("%d", req.Amount))
		valid := req.Amount < 400
		reason := ""
		if valid == false {
			reason = "amount exceeds limit"
		}
		return ValidateResponse{Valid: valid, Reason: reason}, nil

	case MessageForward:
		s.SetTracingSpanAttribute("action", "forward_endpoint")
		s.SetTracingSpanAttribute("hops", fmt.Sprintf("%d", len(req.Hops)))
		s.SetTracingSpanAttribute("total_received", fmt.Sprintf("%d", s.received))
		return PongResponse{
			Payload: fmt.Sprintf("forwarded through %d hops", len(req.Hops)),
			Node:    string(s.Node().Name()),
		}, nil
	}
	return nil, nil
}

func (s *sink) Terminate(reason error) {
	s.Log().Info("tracing sink terminated (received %d): %s", s.received, reason)
}
