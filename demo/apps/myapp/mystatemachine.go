package myapp

import (
	"fmt"
	"time"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

// Order with the states: new, processing, shipped, delivered, canceled

type OrderData struct {
	items     []string
	processed time.Time
	shipped   time.Time
	delivered time.Time
	canceled  time.Time
}

type Order struct {
	act.StateMachine[OrderData]
}

func factoryOrder() gen.ProcessBehavior {
	return &Order{}
}

func (order *Order) Init(args ...any) (act.StateMachineSpec[OrderData], error) {
	spec := act.NewStateMachineSpec(gen.Atom("new"),
		// new
		act.WithStateCallback(gen.Atom("new"), process),
		act.WithStateCallback(gen.Atom("new"), cancel),
		// processing
		act.WithStateCallback(gen.Atom("processing"), ship),
		act.WithStateCallback(gen.Atom("processing"), cancel),
		// shipped
		act.WithStateCallback(gen.Atom("shipped"), deliver),
	)

	return spec, nil
}

type Process struct{}

type Ship struct {
	priority bool
}

type Deliver struct{}

type Cancel struct {
	reason string
}

func process(sm *act.StateMachine[OrderData], message Process) error {
	data := sm.Data()
	if len(data.items) < 1 {
		return fmt.Errorf("can't process order as there are no items added yet")
	}
	sm.Log().Info("processing order...")
	data.processed = time.Now()
	sm.SetData(data)
	sm.SetCurrentState(gen.Atom("processing"))
	return nil
}

func ship(sm *act.StateMachine[OrderData], message Ship) error {
	data := sm.Data()
	sm.Log().Info("shiping order...")
	data.shipped = time.Now()
	sm.SetData(data)
	sm.SetCurrentState(gen.Atom("shipped"))
	return nil
}

func deliver(sm *act.StateMachine[OrderData], message Deliver) error {
	data := sm.Data()
	sm.Log().Info("delivering order...")
	data.delivered = time.Now()
	sm.SetData(data)
	sm.SetCurrentState(gen.Atom("delivered"))
	return nil
}

func cancel(sm *act.StateMachine[OrderData], message Cancel) error {
	data := sm.Data()
	sm.Log().Info("canceling order...")
	data.canceled = time.Now()
	sm.SetData(data)
	sm.SetCurrentState(gen.Atom("canceled"))
	return nil
}
