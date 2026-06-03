package messaging

import (
	"ergo.services/ergo/app"
	"ergo.services/ergo/gen"
)

const (
	appName  gen.Atom = "messaging_scenario"
	poolName gen.Atom = "messaging_pool"
)

func CreateApp() gen.ApplicationBehavior {
	return &messagingApp{}
}

type messagingApp struct {
	app.Application
}

func (a *messagingApp) Load(args ...any) (gen.ApplicationSpec, error) {
	return gen.ApplicationSpec{
		Name:        appName,
		Description: "Messaging scenario: random bursts with variable payload size",
		Mode:        gen.ApplicationModeTemporary,
		Map: map[string]gen.Atom{
			"worker": poolName,
		},
		Network: gen.ApplicationNetwork{
			RegisterTypes: []any{
				MessagePayload{},
				MessageBulkPayload{},
				OrderSide(""),
				OrderTag{},
				TestOrder{},
			},
		},
		Group: []gen.ApplicationMemberSpec{
			{
				Name:    "messaging_sup",
				Factory: factoryMessagingSup,
			},
		},
	}, nil
}
