package tracing

import (
	"ergo.services/ergo/app"
	"ergo.services/ergo/gen"
)

const (
	appName    gen.Atom = "tracing_scenario"
	relayName  gen.Atom = "trace_relay"
	sinkName   gen.Atom = "trace_sink"
	workerName gen.Atom = "trace_worker"
)

func CreateApp() gen.ApplicationBehavior {
	return &tracingApp{}
}

type tracingApp struct {
	app.Application
}

func (a *tracingApp) Load(args ...any) (gen.ApplicationSpec, error) {
	return gen.ApplicationSpec{
		Name:        appName,
		Description: "Tracing scenario: distributed message chains across cluster",
		Mode:        gen.ApplicationModeTemporary,
		Network: gen.ApplicationNetwork{
			RegisterTypes: []any{
				MessagePing{},
				MessagePong{},
				MessageNotify{},
				MessageStatus{},
				PingRequest{},
				PongResponse{},
				ValidateRequest{},
				ValidateResponse{},
				MessageForward{},
			},
		},
		Group: []gen.ApplicationMemberSpec{
			{
				Name:    "tracing_sup",
				Factory: factorySup,
			},
		},
	}, nil
}
